package handler

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OdysseyMomentumExperience/guest-identity-provider/pkg/hydra"
	"github.com/kinbiko/jsonassert"
)

const jsonMediaType = "application/json"
const jsonContentType = jsonMediaType + "; charset=utf-8"

type SubTest struct {
	name           string
	path           string
	method         string
	body           *bytes.Buffer
	expectedStatus int
	expectedData   string
}

// Sort of an integration test,
// since we test the routing and pass some mock data through the hydra client code.
func TestNewHandler(t *testing.T) {
	hydraClient := setupSuite(t)

	subtests := []SubTest{
		{
			name:           "Unknown endpoints should return 404",
			path:           "/foo/bar",
			expectedStatus: 404,
		},
		{
			name:           "Liveness endpoint GET",
			path:           "/",
			expectedStatus: 200,
			expectedData:   "OK",
		},
		{
			name:           "Readiness endpoint GET",
			path:           "/readiness",
			expectedStatus: 200,
			expectedData:   "OIDC ok\nOK",
		},
		{
			name:           "Login endpoint GET, no query params",
			path:           "/v0/guest/login",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			expectedData:   `{"error": "invalid", "message": "<<PRESENCE>>"}`,
		},
		{
			name:           "Login endpoint GET",
			path:           "/v0/guest/login?challenge=foobar",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedData: `{
			"subject": "<<PRESENCE>>",
			"requestURL": "<<PRESENCE>>",
			"display": "<<PRESENCE>>",
			"loginHint": "<<PRESENCE>>",
			"uiLocales": "<<PRESENCE>>"
		    }`,
		},
		{
			name:           "Login endpoint POST invalid input",
			path:           "/v0/guest/login",
			method:         http.MethodPost,
			body:           bytes.NewBufferString(`{"foo": "bar"}`),
			expectedStatus: http.StatusBadRequest,
			expectedData:   `{"error": "invalid", "message": "<<PRESENCE>>"}`,
		},
		{
			name:           "Login endpoint POST",
			path:           "/v0/guest/login",
			method:         http.MethodPost,
			body:           bytes.NewBufferString(`{"challenge": "foobar"}`),
			expectedStatus: http.StatusOK,
			expectedData:   `{"redirect": "<<PRESENCE>>"}`,
		},
		{
			name:           "Consent endpoint GET",
			path:           "/v0/guest/consent",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Consent endpoint POST",
			path:           "/v0/guest/consent",
			method:         http.MethodPost,
			body:           bytes.NewBufferString(`{"challenge": "foobar"}`),
			expectedStatus: http.StatusOK,
			expectedData:   `{"redirect": "<<PRESENCE>>"}`,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.name, func(t *testing.T) {
			request := setupTest(t, &subtest)
			rr := httptest.NewRecorder()
			handler := NewHandler(hydraClient)
			handler.ServeHTTP(rr, request)
			resp := rr.Result()
			if status := resp.StatusCode; status != subtest.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, subtest.expectedStatus)
				data, _ := io.ReadAll(resp.Body)
				t.Fatalf("Response body was %s", data)
			}
			if subtest.expectedData != "" {
				assertBody(t, resp, &subtest.expectedData)
			}
		})
	}
}

func assertBody(t *testing.T, resp *http.Response, expectedData *string) {
	t.Helper()
	ctype := resp.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		t.Fatalf("Invalid Content-Type %s, %s", ctype, err.Error())
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Invalid response body")
	}
	if mediatype == jsonMediaType {
		ja := jsonassert.New(t)
		ja.Assertf(string(data), *expectedData)
	} else {
		if strData := string(data); strData != *expectedData {
			t.Errorf("Handler returned wrong body: got %v want %v",
				strData, *expectedData)
		}
	}
}

func setupSuite(t *testing.T) *hydra.HydraClient {
	hydraClient := setupMockHydra(t)
	return hydraClient
}

func setupTest(t *testing.T, subtest *SubTest) *http.Request {
	request := setupMockRequest(t, subtest)
	return request
}

func setupMockHydra(t *testing.T) *hydra.HydraClient {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/health/alive") {
			w.Header().Set("Content-Type", jsonContentType)
			w.Write([]byte(`{"status": "ok"}`))
		} else if strings.HasSuffix(r.URL.Path, "/login") {
			mockHydraLoginRequest(t, w, r)
		} else if strings.HasSuffix(r.URL.Path, "/login/accept") {
			mockHydraLoginAccept(t, w, r)
		} else if strings.HasSuffix(r.URL.Path, "/consent") {
			mockHydraConsentRequest(t, w, r)
		} else if strings.HasSuffix(r.URL.Path, "/consent/accept") {
			mockHydraConcentAccept(t, w, r)
		} else {
			t.Logf("Unknown call to hydra mock %v", r.URL)
		}
	}))
	hydraClient := hydra.NewHydraClient(server.URL)
	t.Cleanup(server.Close)
	return hydraClient
}

func setupMockRequest(t *testing.T, subtest *SubTest) *http.Request {
	if subtest.body != nil {
		req := httptest.NewRequest(subtest.method, subtest.path, subtest.body)
		req.Header.Set("Content-Type", jsonContentType)
		return req
	}
	return httptest.NewRequest(subtest.method, subtest.path, nil)
}

func mockHydraLoginRequest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", jsonContentType)
	w.Write([]byte(`{
	    "challenge": "foobar",
	    "client": {},
	    "oidc_context": {
	      "acr_values": [],
	      "display": "page",
	      "id_token_hint_claims": {},
	      "login_hint": "barbaz",
	      "ui_locales": ["fr-CA", "fr", "en"]
	    },
	    "request_url": "http://example.com",
	    "requested_access_token_audience": [],
	    "requested_scope": [],
	    "session_id": "",
	    "skip": false,
	    "subject": "subject"
	}`))
}

func mockHydraLoginAccept(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", jsonContentType)
	w.Write([]byte(`{ "redirect_to": "http://example.com" }`))
}

func mockHydraConsentRequest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", jsonContentType)
	w.Write([]byte(`{
	    "requested_access_token_audience" : ["react-client"],
	    "requested_scope" : ["openid"]
	}`))
}
func mockHydraConcentAccept(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", jsonContentType)
	w.Write([]byte(`{"redirect_to": "http://example.com"}`))
}
