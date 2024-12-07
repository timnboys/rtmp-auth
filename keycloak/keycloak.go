package keycloak

import (
	"context"
	"fmt"
	"encoding/base64"
	"encoding/json"
	//"errors"
	"net/http"
	"net/url"
	//"os"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"github.com/timnboys/rtmp-auth/keycl"
)

//Constants

//Global Variables
var client Client                  //Client Object
var realm string                   //realm string from json file
var clientID string                //Client ID string from json file
var clientSecret string            //Client ID sectret from json file
var oauth2Config oauth2.Config     //oath2Config
var provider *oidc.Provider        //oidc provider
var err error                      //generic error object
var keycloakserver string          //keycloak server string passed from app
var server string                  //app server string passed from app
var verifier *oidc.IDTokenVerifier //verifier

type keycloakAuth struct {
	next               http.Handler
	KeycloakURL        *url.URL
	ClientID           string
	ClientSecret       string
	KeycloakRealm      string
	Scope              string
	TokenCookieName    string
	UseAuthHeader      bool
	UserClaimName      string
	UserHeaderName     string
	IgnorePathPrefixes []string
}

func parseUrl(rawUrl string) (*url.URL) {
	if rawUrl == "" {
		return nil
	}
	if !strings.Contains(rawUrl, "://") {
		rawUrl = "https://" + rawUrl
	}
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil
	}
	if !strings.HasPrefix(u.Scheme, "http") {
		//return nil, fmt.Errorf("%v is not a valid scheme", u.Scheme)
		return nil
	}
	return u
}

type state struct {
	RedirectURL string `json:"redirect_url"`
}

func (k *keycloakAuth) redirectToKeycloak(rw http.ResponseWriter, req *http.Request) (cfg *keycl.KeyCloakConfig) {
	scheme := req.Header.Get("X-Forwarded-Proto")
	host := req.Header.Get("X-Forwarded-Host")
	originalURL := fmt.Sprintf("%s://%s%s", scheme, host, req.RequestURI)

	state := state{
		RedirectURL: originalURL,
	}

	stateBytes, _ := json.Marshal(state)
	stateBase64 := base64.StdEncoding.EncodeToString(stateBytes)
	u := parseUrl(cfg.KeyCloakURL)
	k.KeycloakURL = u

	redirectURL := k.KeycloakURL.JoinPath(
		"realms",
		cfg.Realm,
		"protocol",
		"openid-connect",
		"auth",
	)
	redirectURL.RawQuery = url.Values{
		"response_type": {"code"},
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {originalURL},
		"state":         {stateBase64},
		"scope":				 {"email"},
	}.Encode()

	http.Redirect(rw, req, redirectURL.String(), http.StatusTemporaryRedirect)
	return
}

//Init begins keycloak server
func InitKeyCloak(keycloakServer, Server string, cfg keycl.KeyCloakConfig) {
	userLog = GetInstance()
	getKeycloakTOML(cfg)
	keycloakserver = keycloakServer
	server = Server
	ctx := context.Background()
	//Gets the provider for authentication (keycloak)
	//fmt.Println("AUTH URL ", keycloakserver+"/auth/realms/"+cfg.Realm)
	//provider, err = oidc.NewProvider(ctx, keycloakserver+"/realms"+cfg.Realm+"/account/")
	provider, err = oidc.NewProvider(ctx, keycloakserver+"/realms/"+realm)
	if err != nil {
		fmt.Printf("This is an error with regard to the context: %v", err)
	}
	verifier = provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	// Configure an OpenID Connect aware OAuth2 client.
	oauth2Config = oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  server + "/loginCallback",

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

}
