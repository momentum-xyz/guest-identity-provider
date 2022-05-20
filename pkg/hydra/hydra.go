/*
Package to talk to Ory Hydra.

Thin wrapper around Ory's go client, to hide the ugly openapi generated code.
*/
package hydra

import (
	"context"
	"errors"

	client "github.com/ory/hydra-client-go"
)

type HydraClient struct {
	client *client.APIClient
}

func NewHydraClient(adminURL string) *HydraClient {
	hydraCfg := client.NewConfiguration()
	hydraCfg.Servers = client.ServerConfigurations{
		{
			URL: adminURL,
		},
	}
	hydraClient := client.NewAPIClient(hydraCfg)

	return &HydraClient{
		client: hydraClient,
	}
}

// Return Hydra 'alive' status. Should be "ok".
func (h *HydraClient) GetStatus(ctx context.Context) (*string, error) {
	resp, _, err := h.client.MetadataApi.IsAlive(ctx).Execute()
	if err != nil {
		return nil, err
	}
	status, isSet := resp.GetStatusOk()
	if !isSet {
		return nil, errors.New("no status")
	}
	return status, nil
}

// Get login request subject, if user session was still active.
func (h *HydraClient) GetLoginRequest(ctx context.Context, loginChallenge string) (*client.LoginRequest, error) {
	result, _, err := h.client.AdminApi.GetLoginRequest(ctx).LoginChallenge(loginChallenge).Execute()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Accept login request and return URL to redirect the user to.
func (h *HydraClient) AcceptLogin(ctx context.Context, challenge string, subject string) (*string, error) {
	request := client.NewAcceptLoginRequest(subject)
	// For guest sessions only!
	request.SetRemember(true)
	request.SetRememberFor(0) // session cookie
	result, _, err := h.client.AdminApi.AcceptLoginRequest(ctx).LoginChallenge(challenge).AcceptLoginRequest(*request).Execute()
	if err != nil {
		return nil, err
	}
	redirect, isSet := result.GetRedirectToOk()
	if !isSet || *redirect == "" {
		return nil, errors.New("no redirectTo")
	}
	return redirect, nil

}

// Return the requested audience and scope for the concent request
func (h *HydraClient) GetConsent(ctx context.Context, challenge string) (*client.ConsentRequest, error) {
	result, _, err := h.client.AdminApi.GetConsentRequest(ctx).ConsentChallenge(challenge).Execute()
	if err != nil {
		return nil, err
	}
	return result, nil
	/*
		audience, isSet := result.GetRequestedAccessTokenAudienceOk()
		if !isSet {
			return nil, nil, errors.New("no audience requested")
		}
		scope, isSet := result.GetRequestedScopeOk()
		if !isSet {
			return nil, nil, errors.New("no scope requested")
		}
		return audience, scope, nil */
}

func (h *HydraClient) AcceptConsent(ctx context.Context, challenge string, audience []string, scope []string) (*string, error) {
	request := client.NewAcceptConsentRequest()
	// For guest sessions only!
	request.SetRemember(true)
	request.SetRememberFor(0) // And here it means forever
	request.SetGrantAccessTokenAudience(audience)
	request.SetGrantScope(scope)
	session := client.NewConsentRequestSession()
	// Broken type: https://github.com/ory/hydra/issues/3058
	var extraProps = map[string]map[string]interface{}{"guest": {"1": true}}
	// So for now, just give it some 'thruthy' value :/
	session.SetIdToken(
		extraProps,
	)
	request.SetSession(*session)
	result, _, err := h.client.AdminApi.AcceptConsentRequest(ctx).ConsentChallenge(challenge).AcceptConsentRequest(*request).Execute()
	if err != nil {
		return nil, err
	}
	redirect, isSet := result.GetRedirectToOk()
	if !isSet || *redirect == "" {
		return nil, errors.New("no redirectTo")
	}
	return redirect, nil
}
