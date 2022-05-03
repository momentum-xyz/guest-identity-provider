package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/OdysseyMomentumExperience/guest-identity-provider/pkg/hydra"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const hydraClientCtxName = "hydraClient"

func NewHandler(hydraClient *hydra.HydraClient) http.Handler {
	router := chi.NewRouter()

	// Put hydra API client in the context, so we don't need to pass it around
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), hydraClientCtxName, hydraClient)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})

	router.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelCtx := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancelCtx()

		status, err := hydraClient.GetStatus(ctx)
		if err != nil {
			log.Printf("Error getting Hydra status: %v", err.Error())
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(500)
			w.Write([]byte("ERROR"))
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "OIDC %v\n", *status)
		w.Write([]byte("OK"))
	})

	router.Post("/v0/guest/login", loginHandler)
	router.Post("/v0/guest/consent", consentHandler)

	return router
}

// Handle a login for a guest user.
// Normally this would check some authentication method,
// here we just always accept it.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	data := &ChallengeRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	client := getHydraClient(r)
	ctx := r.Context()
	// If we have a subject, it is an 'active' session
	subject, err := client.GetLoginRequest(ctx, data.Challenge)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	var userId string
	if subject != nil && *subject != "" {
		userId = *subject
	} else {
		userId = uuid.NewString() // Guest user, just give them a new ID
	}
	redirectTo, err := client.AcceptLogin(ctx, data.Challenge, userId)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Render(w, r, &RedirectResponse{Redirect: *redirectTo})
}

func getHydraClient(r *http.Request) *hydra.HydraClient {
	client := r.Context().Value(hydraClientCtxName)
	return client.(*hydra.HydraClient)
}

// Handle OIDC consent for a guest user.
// Since these are guest, this just always accept.
func consentHandler(w http.ResponseWriter, r *http.Request) {
	data := &ChallengeRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	client := getHydraClient(r)
	ctx := r.Context()
	audience, scope, err := client.GetConsent(ctx, data.Challenge)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	redirectTo, err := client.AcceptConsent(ctx, data.Challenge, audience, scope)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Render(w, r, &RedirectResponse{Redirect: *redirectTo})
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		ErrorText:      "invalid",
		MessageText:    err.Error(),
	}
}
