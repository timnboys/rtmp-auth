package keycloakmodel

import(
        //"encoding/json"
	//"flag"
	//"io/ioutil"
	//"log"
	//"os"
	//"fmt"
	//"os/signal"
	//"syscall"
	//"time"

	//"github.com/Nerzal/gocloak/v13"
	//"github.com/pelletier/go-toml"
)
import "github.com/timnboys/rtmp-auth/keycl"


/*type KeyCloak struct {
	Client      gocloak.GoCloak // keycloak client
	ClientID     string          // clientId specified in Keycloak
	ClientSecret string          // client secret specified in Keycloak
	Realm        string          // realm specified in Keycloak
	URL          string
}
*/

type KeyCloakModel struct {
	KeyCloak keycl.KeyCloakConfig
}

func NewKeycloak(id keycl.KeyCloakConfig) *KeyCloakModel {
	return &KeyCloakModel {
	KeyCloak: id,
	}
}
