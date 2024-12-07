package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"time"
	"html/template"
	//"path"

	"github.com/timnboys/rtmp-auth/storage"
	"github.com/gorilla/csrf"
	"github.com/timnboys/rtmp-auth/store"
	"github.com/timnboys/rtmp-auth/keycl"
	//"github.com/timnboys/rtmp-auth/assets"
	"context"
	"github.com/zemirco/keycloak"
    	"golang.org/x/oauth2"
	//"github.com/Nerzal/gocloak/v13"
)

type handleFunc func(http.ResponseWriter, *http.Request)

var durationRegex = regexp.MustCompile(`P([\d\.]+Y)?([\d\.]+M)?([\d\.]+D)?T?([\d\.]+H)?([\d\.]+M)?([\d\.]+?S)?`)

func parseDurationPart(value string, unit time.Duration) time.Duration {
	if len(value) != 0 {
		if parsed, err := strconv.ParseFloat(value[:len(value)-1], 64); err == nil {
			return time.Duration(float64(unit) * parsed)
		}
	}
	return 0
}

// Parse expiration time
func parseExpiry(str string) *int64 {
	// Allow empty string for "never"
	if str == "" {
		never := int64(-1)
		return &never
	}

	// Try to parse as ISO8601 duration
	matches := durationRegex.FindStringSubmatch(str)
	if matches != nil {
		years := parseDurationPart(matches[1], time.Hour*24*365)
		months := parseDurationPart(matches[2], time.Hour*24*30)
		days := parseDurationPart(matches[3], time.Hour*24)
		hours := parseDurationPart(matches[4], time.Hour)
		minutes := parseDurationPart(matches[5], time.Second*60)
		seconds := parseDurationPart(matches[6], time.Second)
		d := time.Duration(years + months + days + hours + minutes + seconds)
		if d == 0 {
			return nil
		}

		expiry := time.Now().Add(d).Unix()
		return &expiry
	}

	// Try to parse as absolute time
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return nil
	}
	expiry := t.Unix()
	return &expiry
}

type SRSPublish struct {
	Action string `json:"action"`
	IP     string `json:"ip"`
	VHost  string `json:"vhost"`
	App    string `json:"app"`
	Url    string `json:"tcUrl"`
	Stream string `json:"stream"`
	Param  string `json:"param"`
}

func handleSRSPublish(r *http.Request) (app string, name string, auth string, action string, err error) {
	defer r.Body.Close()
	var publish SRSPublish
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&publish)
	if err != nil {
		return
	}

	// skip question mark
	if len(publish.Param) > 0 {
		publish.Param = publish.Param[1:]
	}

	val, err := url.ParseQuery(publish.Param)
	if err != nil {
		return
	}
	app = publish.App
	name = publish.Stream
	auth = val.Get("auth")
	action = publish.Action
	return
}

func handleNginxPublish(r *http.Request) (app string, name string, auth string, action string, err error) {
	err = r.ParseForm()
	if err != nil {
		return
	}

	app = r.PostForm.Get("app")
	name = r.PostForm.Get("name")
	auth = r.PostForm.Get("auth")
	action = r.PostForm.Get("call")
	return
}

func PublishHandler(store *store.Store) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var app string
		var name string
		var auth string
		var action string
		var err error

		if r.Header.Get("Content-Type") == "application/json" {
			// SRS publish handler
			app, name, auth, action, err = handleSRSPublish(r)
			if action != "on_publish" {
				err = fmt.Errorf("invalid action %s", action)
			}
		} else {
			// Form DATA from nginx-rtmp/srtrelay
			app, name, auth, action, err = handleNginxPublish(r)
			log.Println("publish action", action)

			// only apply auth for publish
			if action != "publish" {
				return
			}
		}
		log.Println(app, name, auth, err)
		if err != nil {
			log.Println("Failed to parse publish data:", err)
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("publish %s/%s auth: '%s'\n", app, name, auth)

		success, id := store.Auth(app, name, auth)
		if !success {
			log.Printf("Publish %s %s/%s unauthorized\n", id, app, name)
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		store.SetActive(id)
		log.Printf("Publish %s %s/%s ok\n", id, app, name)

		// SRS needs zero response
		w.Write([]byte("0"))
	}
}

func UnpublishHandler(store *store.Store) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var app string
		var name string
		var action string
		var err error

		if r.Header.Get("Content-Type") == "application/json" {
			// SRS publish handler
			app, name, _, action, err = handleSRSPublish(r)
			if action != "on_unpublish" {
				err = fmt.Errorf("invalid action %s", action)
			}
		} else {
			// Form DATA from nginx-rtmp/srtrelay
			app, name, _, action, err = handleNginxPublish(r)
			log.Println("unpublish action", action)
			// ignore actions except unpublish
			if action != "unpublish" {
				return
			}
		}

		if err != nil {
			log.Println("Failed to parse unpublish data:", err)
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}
		store.SetInactive(app, name)
		log.Printf("Unpublish %s/%s ok\n", app, name)

		// SRS needs zero response
		w.Write([]byte("0"))
	}
}

func FormHandler(store *store.Store, config ServerConfig) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errs []error
		state, err := store.Get()
		if err != nil {
			errs = append(errs, err)
		}

		sort.SliceStable(state.Streams, func(i, j int) bool {
			return state.Streams[i].Name < state.Streams[j].Name
		})

		data := TemplateData{
			State:        state,
			Config:       config,
			CsrfTemplate: csrf.TemplateField(r),
			Errors:       errs,
		}
		err = templates.ExecuteTemplate(w, "form.html", data)
		if err != nil {
			log.Println("Template failed", err)
		}
	}
}

func LoginFormHandler(store *store.Store, config ServerConfig, cfg keycl.KeyCloakConfig) handleFunc {
        return func(w http.ResponseWriter, r *http.Request) {
		/*
                var errs []error
                state, err := store.Get()
                if err != nil {
                        errs = append(errs, err)
                }

                sort.SliceStable(state.Streams, func(i, j int) bool {
                        return state.Streams[i].Name < state.Streams[j].Name
                })
		err = templates.ExecuteTemplate(w, "public/logon.html", nil)
		if err != nil {
                        log.Println("Template failed", err)
                }
		//w.Write([]byte("The Login Page should have appeared..."))
		*/
		tmpl, err := template.ParseFiles("public/login.html")
                        if err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                        return
                        }
               tmpl.Execute(w, nil)
		/*
                data := LoginPageTemplateData{
                        State:        state,
                        Config:       config,
                        CsrfTemplate: csrf.TemplateField(r),
                        Errors:       errs,
                }
		fp := path.Join("public", "index.html")
   		tmpl, err := template.ParseFiles(fp)
    		if err != nil {
        	http.Error(w, err.Error(), http.StatusInternalServerError)
        	return
    		}

		if err := tmpl.Execute(w, fp); err != nil {
        	http.Error(w, err.Error(), http.StatusInternalServerError)
    		}
                err = templates.ExecuteTemplate(w, "logon.html", data)
                if err != nil {
                        log.Println("Template failed", err)
                }
		*/
        }
}

// LoginPage is the handler for the login page.
func LoginHandler(store *store.Store, config keycl.KeyCloakConfig) handleFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        username := r.FormValue("username")
        password := r.FormValue("password")

        // Perform authentication logic here (e.g., check against a database).
	//ctx := context.Background()
	// create your oauth configuration
    keyclconfig := oauth2.Config{
        ClientID: "admin-cli",
        Endpoint: oauth2.Endpoint{
            TokenURL: config.KeyCloakTokenURL,
        },
    }

    // get a valid token from keycloak
    token, err := keyclconfig.PasswordCredentialsToken(context.Background(), username, password)
    if err != nil {
			state, err := store.Get()
			data := LoginPageTemplateData{
				State:        state,
				//Config:       config,
				CsrfTemplate: csrf.TemplateField(r),
				//Errors:       err,
			}
			//err = w.Write([]byte())
			err = templates.ExecuteTemplate(w, "public/login.html", data)
			if err != nil {
				log.Println("Template failed", err)
				fmt.Println("Template failed", err)
			}
    }

	// create a new http client
	httpClient := keyclconfig.Client(context.Background(), token)
	// use the http client to create a Keycloak instance
	kc, err := keycloak.NewKeycloak(httpClient, config.KeyCloakURL)
	if err != nil {
		panic(err)
		fmt.Println(err)
	}

	// then use this instance to make requests to the API
	fmt.Println(kc)// use the http client to create a Keycloak instance

    // create a new http client that uses the token on every request
    	//client := keyclconfig.Client(context.Background(), token)
	// create a new keycloak instance and provide the http client
	/*
	k, err := keycloak.NewKeycloak(client, config.KeyCloakURL)
    	if err != nil {
        fmt.Fprintf(w, "Invalid credentials. Please try again.")
        tmpl, err := template.ParseFiles("templates/login.html")
        if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
        }
        tmpl.Execute(w, nil)
    	}
	//keycl.KeyCloakConfig.Client := k
	

        // Invalid credentials, show the login page with an error message.
        fmt.Fprintf(w, "Invalid credentials. Please try again.")
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
	tmpl.Execute(w, nil)
    }
	// get the client to have all properties
	client, res, err = kc.Clients.Get(ctx, realm, id)
	if err != nil {
		//panic(err)
	fmt.Fprintf(w, "Invalid credentials. Please try again.")
        tmpl, err := template.ParseFiles("templates/login.html")
        if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
        }
        tmpl.Execute(w, nil)
	}
*/
    // If not a POST request, serve the login page template.
    //tmpl, err := template.ParseFiles("login.html")
    //if err != nil {
    //    http.Error(w, err.Error(), http.StatusInternalServerError)
    //    return
    }
    //tmpl.Execute(w, nil)
    }
}

func AddHandler(store *store.Store, config ServerConfig) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errs []error

		expiry := parseExpiry(r.PostFormValue("auth_expire"))
		if expiry == nil {
			errs = append(errs, fmt.Errorf("invalid auth expiry: '%v'", r.PostFormValue("auth_expire")))
		}

		name := r.PostFormValue("name")
		if len(name) == 0 {
			errs = append(errs, fmt.Errorf("stream name must be set"))
		}

		// TODO: more validation
		if len(errs) == 0 {
			stream := &storage.Stream{
				Name:        name,
				Application: r.PostFormValue("application"),
				AuthKey:     r.PostFormValue("auth_key"),
				AuthExpire:  *expiry,
				Notes:       r.PostFormValue("notes"),
			}

			err := store.AddStream(stream)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to add stream: %w", err))
			} else {
				http.Redirect(w, r, config.Prefix, http.StatusSeeOther)
			}
		}

		state, err := store.Get()
		if err != nil {
			errs = append(errs, err)
		}
		data := LoginPageTemplateData{
			State:        state,
			Config:       config,
			CsrfTemplate: csrf.TemplateField(r),
			Errors:       errs,
		}
		err = templates.ExecuteTemplate(w, "form.html", data)
		if err != nil {
			log.Println("Template failed", err)
		}
	}
}

func RemoveHandler(store *store.Store, config ServerConfig) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errs []error
		id := r.PostFormValue("id")

		err := store.RemoveStream(id)
		if err != nil {
			log.Println(err)
			errs = append(errs, fmt.Errorf("failed to remove stream: %w", err))
			state, err := store.Get()
			if err != nil {
				errs = append(errs, err)
			}
			data := TemplateData{
				State:        state,
				Config:       config,
				CsrfTemplate: csrf.TemplateField(r),
				Errors:       errs,
			}
			err = templates.ExecuteTemplate(w, "form.html", data)
			if err != nil {
				log.Println("Template failed", err)
			}
		} else {
			http.Redirect(w, r, config.Prefix, http.StatusSeeOther)
		}
	}
}

func BlockHandler(store *store.Store, config ServerConfig) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errs []error
		id := r.PostFormValue("id")
		last, _ := strconv.ParseBool(r.PostFormValue("blocked"))
		new := false
		action := "unblock"
		if !last {
			new = true
			action = "block"
		}

		// Get Application/Name for stream id
		var app, name string
		state, err := store.Get()
		if err != nil {
			errs = append(errs, err)
		}
		for _, stream := range state.Streams {
			if stream.Id == id {
				app = stream.Application
				name = stream.Name
			}
		}

		err = store.SetBlocked(id, new)
		if err != nil {
			log.Println(err)
			errs = append(errs, fmt.Errorf("failed to %v stream %v (%v/%v)", action, id, app, name))
		}
		log.Printf("%ved Stream %v (%v/%v)", action, id, app, name)
		if len(errs) > 0 {
			data := TemplateData{
				State:        state,
				Config:       config,
				CsrfTemplate: csrf.TemplateField(r),
				Errors:       errs,
			}
			err = templates.ExecuteTemplate(w, "form.html", data)
			if err != nil {
				log.Println("Template failed", err)
			}
		} else {
			http.Redirect(w, r, config.Prefix, http.StatusSeeOther)
		}
	}
}
