# Momentum guest account identity provider

A http API that implements [Hydra](https://www.ory.sh/docs/hydra) 
login and consent for use by guest logins.

Guest login, so don't expect security :)
It is up to the backend to use these access token with appropriate authorization.

Loosely based on [web3-identity-provider](https://github.com/OdysseyMomentumExperience/web3-identity-provider)


## Building

```console
make clean build
```

Executable in place in `output/guest-identity-provider`.

Or build as container image:

```console
make clean container-image
```

Container image is then in your docker registry as `guest-identity-provider:develop`
and as exported file in `output/container.img.tar`


## Running

This application depends on having Hydra running and direct access to it's admin API.

It can be configured by a yaml config file and/or environment variables.

| env | Description | Default |
| --- | --- | --- |
| `GUEST_IDP_HOST` | The host to bind to | localhost |
| `GUEST_IDP_PORT` | The port to bind to | 4000 |
| `HYDRA_ADMIN_URL` | Absolute URL to the hydra admin service | http://localhost:4445 |
| `CONFIG_FILE` | Name/path of a YAML config file to read | config.yaml |

And/or using a YAML config file, see `config.example.yaml`.

The `guest-identity-provider` executable does not accept any arguments. It will just start listening on the configured host and port.


## Usage

HTTP API endpoints:

|Method | path | Description | input | output |
| --- | --- | --- | --- | --- |
|GET | `/v0/guest/login`   | Get login session info | `?challenge=…` | `{"subject": "…", "requestURL: "…", "display": "…", "loginHint": "…", "uiLocales": […]}` "
|POST | `/v0/guest/login`   | Accept login | `{"challenge": "…"}` | `{"redirect": "https://…"}`
|POST | `/v0/guest/consent` | Accept consent | `{"challenge": "…"}` | `{"redirect": "https://…"}`


## Testing

Project has some basic golang tests.

```console
make test
```

And see `scripts/test.py` for a e2e test.
