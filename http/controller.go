package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	//"github.com/pelletier/go-toml"
	//"github.com/Nerzal/gocloak/v13"	
        
)
import "github.com/timnboys/rtmp-auth/keycl"

type doc struct {
	Id   string    `json:"id"`
	Num  string    `json:"num"`
	Date time.Time `json:"date"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

type controller struct {
   keycloak keycl.KeyCloakConfig
}

func newController(conf keycl.KeyCloakConfig) *controller {
	return &controller{
		keycloak: conf,
	}
}

func (c *controller) login(w http.ResponseWriter, r *http.Request) {

	rq := &loginRequest{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(rq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jwt, err := c.keycloak.Client.Login(context.Background(),
		c.keycloak.ClientID,
		c.keycloak.ClientSecret,
		c.keycloak.Realm,
		rq.Username,
		rq.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	rs := &loginResponse{
		AccessToken:  jwt.AccessToken,
		RefreshToken: jwt.RefreshToken,
		ExpiresIn:    jwt.ExpiresIn,
	}

	rsJs, _ := json.Marshal(rs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(rsJs)
}
