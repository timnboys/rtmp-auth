package keycloak

import (
	"context"
	//"encoding/json"
	//"bytes"
	"fmt"
	//"io/ioutil"
	"net/http"
	"log"

	"golang.org/x/oauth2"
)

var oauthStateString string //randomly generated state string
var token *oauth2.Token     //token for keycloak

//HandleLogin is the keycloak login funtion
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	//create a random string for oath2 verification
	oauthStateString = randSeq(20)
	//Uses random gnerated string to verify keyclock security
	url := oauth2Config.AuthCodeURL(oauthStateString)
	//fmt.Printf("Detected HandleLogin called, redirecting to URL '%s'",url) 
	//redirects to loginCallback
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

//HandleLoginCallback is a fuction that verifies login success and forwards to index
func HandleLoginCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	//Checks that the strings are in a consistent state
	if state != oauthStateString {
		log.Println("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		HandleLogin(w,r)
		//http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	//Gets the code from keycloak
	code := r.FormValue("code")
	//Exchanges code for token
	token, err = oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Println("Code exchange failed with '%v'\n", err)
		HandleLogin(w,r)
		//http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	client := &http.Client{}
	url := keycloakserver + "/realms/" + realm + "/protocol/openid-connect/userinfo"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
        w.Header().Set("Authorization", "Bearer "+token.AccessToken)
	//Sends the token to get user info
	response, err := client.Do(req)

	//Checks if token and authentication were successful
	if err != nil || response.Status != "200 OK" {
		//forwards back to login if not successful
		//http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		fmt.Errorf("Code exchange failed with '%v'\n", err)
		log.Println("Code exchange failed with '%v'\n", err)
		HandleLogin(w,r)
	} else {
		//body, _ := ioutil.ReadAll(response.Body)
		//var f interface{}
		//json.Unmarshal(body, &f)
		//m := f.(map[string]interface{})
		//username := m["preferred_username"].(string)
		//forwards to index if login sucessful
		//logAction(username, actionLogin, "")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	return
}

//AuthMiddleware is a middlefuntion that verifies authentication before each redirect
func AuthMiddleware(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//If running unit tests skip authentication (temp)
		client := &http.Client{}
		url := keycloakserver + "/realms/" + realm + "/protocol/openid-connect/userinfo"
		req, _ := http.NewRequest("GET", url, nil)
		if token == nil && r.URL.Path != "/loginCallback" {
			log.Println("Token Nil")
			HandleLogin(w,r)
			//http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}
		if r.URL.Path == "/loginCallback" {
			//log.Println("Token Nil and LoginCallback")
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/logout" {
			log.Println("Token Nil and Logout")
			next.ServeHTTP(w, r)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		//Check if token is still valid
		response, err := client.Do(req)
		if err != nil || response.Status != "200 OK" {
			//Go to login if token is no longer valid
			//HandleLogin(w,r)
			//log.Println("Error is ",err)
			//log.Println("Response Status is ",response.Status)
			//http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			HandleLogin(w,r)
		} else {
			//body, _ := ioutil.ReadAll(response.Body)
			//var f interface{}
			//json.Unmarshal(body, &f)
			//m := f.(map[string]interface{})
			//username := m["preferred_username"].(string)
			//Go to redirect if token is still valid
			//logAction(username, actionPageAccess, r.RequestURI)
			next.ServeHTTP(w, r)
		}
	})
	//return function for page handling
	return handler
}

//Logout logs the user out
func Logout(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	url := keycloakserver + "/realms/" + realm + "/protocol/openid-connect/userinfo"
	req, _ := http.NewRequest("GET", url, nil)
	if token == nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	//Check if token is still valid
	response, err := client.Do(req)
	if response.Status == "200 OK" && err == nil {
		//body, _ := ioutil.ReadAll(response.Body)
		//var f interface{}
		//json.Unmarshal(body, &f)
		//m := f.(map[string]interface{})
		//username := m["preferred_username"].(string)
		//Go to redirect if token is still valid
		//logAction(username, actionLogout, "")
	}
	//Makes the logout page redirect to login page
	URI := server + "/"
	//Logout using endpoint and redirect to login page
	http.Redirect(w, r, keycloakserver+"/realms/"+realm+"/protocol/openid-connect/logout?redirect_uri="+URI, http.StatusTemporaryRedirect)
}

