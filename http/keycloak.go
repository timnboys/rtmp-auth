package main

import "github.com/Nerzal/gocloak/v13"

type keycloak struct {
	gocloak      gocloak.GoCloak // keycloak client
	clientId     string          // clientId specified in Keycloak
	clientSecret string          // client secret specified in Keycloak
	realm        string          // realm specified in Keycloak
}

func newKeycloak() *keycloak {
	return &keycloak{
		gocloak:      gocloak.NewClient(config.KeyCloak),
		clientId:     config.ClientID,
		clientSecret: config.ClientSecret,
		realm:        config.Realm,
	}
}