package api

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/OpenFactorioServerManager/factorio-server-manager/bootstrap"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

var csrfMiddleware func(http.Handler) http.Handler

// SetupCSRF initializes CSRF protection middleware
func SetupCSRF() {
	config := bootstrap.GetConfig()

	// Decode the cookie encryption key to use for CSRF
	csrfKey, err := base64.StdEncoding.DecodeString(config.CookieEncryptionKey)
	if err != nil {
		log.Printf("Error decoding base64 cookie encryption key for CSRF: %s", err)
		panic(err)
	}

	// Use only first 32 bytes for CSRF key
	if len(csrfKey) > 32 {
		csrfKey = csrfKey[:32]
	}

	csrfMiddleware = csrf.Protect(
		csrfKey,
		csrf.Secure(config.Secure),
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.HttpOnly(true),
		csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)),
	)
}

// GetCSRFMiddleware returns the CSRF middleware
func GetCSRFMiddleware() func(http.Handler) http.Handler {
	return csrfMiddleware
}

// csrfErrorHandler handles CSRF validation errors
func csrfErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("CSRF validation failed for %s %s", r.Method, r.URL.Path)
	http.Error(w, "CSRF token validation failed", http.StatusForbidden)
}

// CSRFTokenHandler returns the CSRF token for the current session
func CSRFTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	token := csrf.Token(r)
	WriteResponse(w, map[string]string{"token": token})
}

// ApplyCSRFToRouter applies CSRF protection to the router
func ApplyCSRFToRouter(router *mux.Router) {
	if csrfMiddleware != nil {
		router.Use(csrfMiddleware)
	}
}
