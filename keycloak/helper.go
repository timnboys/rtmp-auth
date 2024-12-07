package keycloak

import (
	//"encoding/json"
	"fmt"
	//"io/ioutil"
	"math/rand"
	//"os"
	"github.com/timnboys/rtmp-auth/keycl"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//randSeq generates a random string of letters of the given length (Helper function)
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

//gets the keycloak JSON informtion from the json file in the json directrory of the users app
func getKeycloakTOML(cfg keycl.KeyCloakConfig) {
	if cfg.ClientID == "" || cfg.Realm == "" || cfg.ClientSecret == "" {
		fmt.Printf("Error reading keycloak file")
	}
	realm = cfg.Realm
	clientID = cfg.ClientID
	clientSecret = cfg.ClientSecret
}
