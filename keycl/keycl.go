package keycl

import (
"github.com/Nerzal/gocloak/v13"
//"github.com/pelletier/go-toml"
)

/*
type KeyCloakConfig struct {
	Client      gocloak.GoCloak // keycloak client
	ClientID     string          // clientId specified in Keycloak
	ClientSecret string          // client secret specified in Keycloak
	KeyCloakURL  string
	Realm        string          // realm specified in Keycloak
}
*/


type KeyCloakConfig struct {
	Client       gocloak.GoCloak // keycloak client
        ClientID     string `toml:"kc-oauth-cl-id"`
        ClientSecret string `toml:"kc-oauth-cl-secret"`
        KeyCloakURL  string `toml:"keycloakurl"`
	FrontendAppAddress  string `toml:"appfrontendurl"`
        Realm        string `toml:"keycloakrealm"`
}
