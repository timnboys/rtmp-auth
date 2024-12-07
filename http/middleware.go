package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"html/template"
	//"os"
	"github.com/Nerzal/gocloak/v13"
	//"github.com/timnboys/rtmp-auth/storage"
	//"github.com/gorilla/csrf"
	//"github.com/timnboys/rtmp-auth/store"
	//"gopkg.in/unrolled/render.v1"
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
		//r := render.New(render.Options{})
		// try to extract Authorization parameter from the HTTP header
		token := r.Header.Get("Authorization")

		if token == "" {
			//http.Redirect(w, r, "https://"+r.Host+"/public/login.html", http.StatusMovedPermanently)
			//http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			// If not a POST request, serve the login page template.
    			tmpl, err := template.ParseFiles("public/login.html")
    			if err != nil {
        		http.Error(w, err.Error(), http.StatusInternalServerError)
        		return
    			}
    			tmpl.Execute(w, nil)
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

		jwt, _, err := auth.keycloak.Client.DecodeAccessToken(context.Background(), token, auth.keycloak.Realm)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid or malformed token: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		//jwtj, _ := json.Marshal(jwt)

		// check if the token isn't expired and valid
		if !*result.Active {
			http.Error(w, "Invalid or expired Token", http.StatusUnauthorized)
			return
		}

		fmt.Println("JWT ", jwt)
		fmt.Println("Token ", token)

		/*
		rs := &loginResponse{
		AccessToken:  jwt.AccessToken,
		RefreshToken: jwt.RefreshToken,
		ExpiresIn:    jwt.ExpiresIn,
		}

		rsJs, _ := json.Marshal(rs)
		*/

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Authorization", "Bearer " + token)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

type LoginResponse struct {
   AccessToken string `json:”access_token”`
   Title string `json:”Title”`
   Description string `json:”Description”`
}


func (auth *keyCloakMiddleware) Protect(next http.Handler) http.Handler {
   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       client := gocloak.NewClient(auth.keycloak.KeyCloakURL)
       authHeader := r.Header.Get("Authorization")
       if len(authHeader) < 1 {
          w.WriteHeader(401)
         json.NewEncoder(w).Encode("Error, Unauthorized")
         return
      }
     accessToken := strings.Split(authHeader," ")[1]
     rptResult, err := client.RetrospectToken(r.Context(),
     accessToken, auth.keycloak.ClientID, auth.keycloak.ClientSecret, auth.keycloak.Realm)
     if err != nil{
      w.WriteHeader(400)
      json.NewEncoder(w).Encode(err.Error())
      return
     }
    isTokenValid := *rptResult.Active
    if !isTokenValid {
      w.WriteHeader(401)
      json.NewEncoder(w).Encode("Error, Unauthorized")
      return
    }
    next.ServeHTTP(w, r)
  })
}
