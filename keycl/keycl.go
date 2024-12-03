package keycl

import (
"github.com/Nerzal/gocloak/v13"
)

/*type keycloak struct {
	gocloak      gocloak.GoCloak // keycloak client
	clientId     string          // clientId specified in Keycloak
	clientSecret string          // client secret specified in Keycloak
	KeyCloakURL  string
	realm        string          // realm specified in Keycloak
}
*/

type KeyCloakConfig struct {
	Client       gocloak.GoCloak // keycloak client
        ClientID     string `toml:"kc-oauth-cl-id"`
        ClientSecret string `toml:"kc-oauth-cl-secret"`
        KeyCloakURL  string `toml:"keycloakurl"`
        Realm        string `toml:"keycloakrealm"`
}
