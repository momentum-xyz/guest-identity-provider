"""
Little script that runs a OIDC/Oauth2 flow with the guest-identity-provider.
So kind of like an e2e test.

Prerequisites:
    - python (3.10 currently)
    - Running (or port forwarded) hydra on oidc.localhost
    - Configured a hydra client, with its client_id in OIDC_CLIENT env var
    - Running guest-identity-provider on guest-idp.localhost:4000
    - or change all the global constants in this file.

Usage:
Change directory to where this script lives.

$ pipenv shell
$ python test.py
"""
import json
import os
from pprint import pprint
from sys import stderr
from urllib.parse import parse_qs, urlparse

import httpx
import jwt
import trio
from authlib.integrations.httpx_client import AsyncOAuth2Client

OIDC_CLIENT = os.environ.get("OIDC_CLIENT", "f2f9cd70-fdd9-4e88-aba3-aff68357759e")
OIDC_SECRET = None
OIDC_SERVER = "oidc.localhost"
OIDC_URL = f"http://{OIDC_SERVER}"
OIDC_DISCO_PATH = "/.well-known/openid-configuration"
OIDC_DISCOVERY = f"{OIDC_URL}{OIDC_DISCO_PATH}"

REDIRECT_URI = "http://localhost:3000/oidc/guest/callback"

_TMP_OIDC = "/tmp/py_oidc.json"


async def main():
    oidc_disco = await get_oidc_discovery()
    authorization_endpoint = oidc_disco["authorization_endpoint"]
    token_endpoint = oidc_disco["token_endpoint"]
    print("Discovered OIDC:")
    print(f"{authorization_endpoint=}")
    print(f"{token_endpoint=}")
    print(f'{oidc_disco["scopes_supported"]=}')
    print()
    scope = "openid"
    await oidc_auth(
        OIDC_CLIENT,
        OIDC_SECRET,
        authorization_endpoint,
        token_endpoint,
        scope,
        redirect_uri=REDIRECT_URI,
    )


async def oidc_auth(
    client_id,
    client_secret,
    authorization_endpoint,
    token_endpoint,
    scope,
    redirect_uri,
):

    # Mock a user's browser session (cookies are stored)
    with httpx.Client() as user_client:
        print("---app---")
        print("App initiates an oauth2 flow")  # hopefully with a OIDC client library.
        oidc_client = AsyncOAuth2Client(
            client_id=client_id,
            client_secret=client_secret,
            scope=scope,
            redirect_uri=redirect_uri,
        )
        authorization_url, state = oidc_client.create_authorization_url(
            authorization_endpoint,
            audience="react-client",
        )
        print(f"{authorization_url=}")
        print(f"{state=}")

        print("App's oidc client redirects browser to authorization URL")
        print("---oidc---")
        r = user_client.get(authorization_url)
        assert r.status_code == 302
        print("OIDC implementation redirects browser to login URL")
        login_url = r.headers["location"]
        print(f"{login_url=}")

        print("---login app---")
        print("Login app app gets login challenge")
        login_challenge = parse_qs(urlparse(login_url).query).get("login_challenge")[0]
        print(f"{login_challenge=}")

        print("Login app send challenge to idp service")
        r = user_client.post(
            "http://guest-idp.localhost:4000/v0/guest/login",
            json={"challenge": login_challenge},
        )
        assert r.status_code == 200
        print("idp sevice sends back a redirect URL")
        login_accepted_url = r.json()["redirect"]
        print(f"{login_accepted_url=}")
        print("Login app redirects browser to URL")

        print("---oidc---")
        r = user_client.get(login_accepted_url)
        print("OIDC implementation redirects browser to consent URL")
        consent_url = r.headers["location"]
        print(f"{consent_url=}")

        print("---login app---")
        print("Login app gets consent challenge")
        consent_challenge = parse_qs(urlparse(consent_url).query).get(
            "consent_challenge"
        )[0]
        print(f"{consent_challenge}")

        print("Login app send challenge to idp service")
        r = user_client.post(
            "http://guest-idp.localhost:4000/v0/guest/consent",
            json={"challenge": consent_challenge},
        )
        assert r.status_code == 200
        print("idp sevice sends back a redirect URL")
        consent_accepted_url = r.json()["redirect"]
        print(f"{consent_accepted_url=}")
        print("Login app redirects browser to URL")

        print("---oidc---")
        r = user_client.get(consent_accepted_url)
        print("OIDC implementation redirects browser to callback URL")
        callback_url = r.headers["location"]
        print(f"{callback_url=}")

        print("---app---")
        print("App exchanges code for tokens")
        token = await oidc_client.fetch_token(
            token_endpoint, authorization_response=callback_url
        )
        access_token = token["access_token"]
        id_token = token["id_token"]
        # print(f'{access_token=}')
        access_token_decoded = jwt.decode(
            access_token, options={"verify_signature": False}
        )
        print("Decoded access_token:")
        pprint(access_token_decoded)

        id_token_decoded = jwt.decode(
            id_token, options={"verify_signature": False}
        )
        print("Decoded id_token:")
        pprint(id_token_decoded)


async def get_oidc_discovery():
    try:
        async with await trio.open_file(_TMP_OIDC, encoding="utf-8") as f:
            cached_content = json.loads(await f.read())
            return cached_content
    except FileNotFoundError as e:
        async with httpx.AsyncClient(http2=True) as client:
            http_response = await client.get(OIDC_DISCOVERY)
            response = http_response.json()
        async with await trio.open_file(_TMP_OIDC, "w", encoding="utf-8") as f:
            await f.write(json.dumps(response, indent=2))
        return response


trio.run(main)
