package http

import (
	"context"
	//"encoding/json"
	"fmt"
	"net/http"
	"strings"
)
import "github.com/timnboys/rtmp-auth/keycl"

type keyCloakMiddleware struct {
	keycloak keycl.KeyCloakConfig
}

func newMiddleware(KeyCloak keycl.KeyCloakConfig) *keyCloakMiddleware {
	return &keyCloakMiddleware{keycloak: KeyCloak}
}

func (auth *keyCloakMiddleware) extractBearerToken(token string) string {
	return strings.Replace(token, "Bearer ", "", 1)
}

func (auth *keyCloakMiddleware) verifyToken(next http.Handler) http.Handler {

	f := func(w http.ResponseWriter, r *http.Request) {

		// try to extract Authorization parameter from the HTTP header
		token := r.Header.Get("Authorization")

		if token == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// extract Bearer token
		token = auth.extractBearerToken(token)

		if token == "" {
			http.Error(w, "Bearer Token missing", http.StatusUnauthorized)
			return
		}

		//// call Keycloak API to verify the access token
		result, err := auth.keycloak.Client.RetrospectToken(context.Background(), token, auth.keycloak.ClientID, auth.keycloak.ClientSecret, auth.keycloak.Realm)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid or malformed token: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		//jwt, _, err := auth.keycloak.Client.DecodeAccessToken(context.Background(), token, auth.keycloak.Realm)
		//if err != nil {
		//	http.Error(w, fmt.Sprintf("Invalid or malformed token: %s", err.Error()), http.StatusUnauthorized)
		//	return
		//}

		//jwtj, _ := json.Marshal(jwt)

		// check if the token isn't expired and valid
		if !*result.Active {
			http.Error(w, "Invalid or expired Token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
