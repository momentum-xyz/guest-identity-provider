package handler

import (
	"errors"
	"net/http"
)

// Input for the endpoints
type ChallengeRequest struct {
	Challenge string `json:"challenge"`
}

func (c *ChallengeRequest) Bind(r *http.Request) error {
	if c.Challenge == "" {
		return errors.New("missing required challenge field.")
	}
	return nil
}

// Response for client side redirects.
type RedirectResponse struct {
	Redirect string `json:"redirect"`
}

func (rr *RedirectResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Response for getting login information.
type LoginInfoResponse struct {
	Subject    string   `json:"subject,omitempty"`
	RequestURL string   `json:"requestURL"`
	Display    string   `json:"display,omitempty"`
	LoginHint  string   `json:"loginHint,omitempty"`
	UILocales  []string `json:"uiLocales,omitempty"`
}

func (l *LoginInfoResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type ErrResponse struct {
	Err            error `json:"-"`
	HTTPStatusCode int   `json:"-"`

	ErrorText   string `json:"error"`
	MessageText string `json:"message,omitempty"`
}
