package http

import (
	"context"
	"log"
	"sync"
	"time"
	//"fmt"
	//"html/template"
	//"os"

	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	_ "github.com/timnboys/rtmp-auth/statik"
	"github.com/timnboys/rtmp-auth/store"
	"github.com/timnboys/rtmp-auth/keycl"
	"github.com/timnboys/rtmp-auth/keycloak"
	//"gopkg.in/unrolled/render.v1"
)

type ServerConfig struct {
	Applications []string `toml:"applications"`
	Prefix       string   `toml:"prefix"`
	Insecure     bool     `toml:"insecure"`
}

type Frontend struct {
	server *http.Server
	done   sync.WaitGroup
}

func NewFrontend(address string, config ServerConfig, cfg keycl.KeyCloakConfig, store *store.Store) *Frontend {
	state, err := store.Get()
	if err != nil {
		log.Fatal("get", err)
	}
	CSRF := csrf.Protect(state.Secret, csrf.Secure(!config.Insecure))
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	//r := render.New(render.Options{})
	router := mux.NewRouter()
	sub := router.PathPrefix(config.Prefix).Subrouter()
	//noAuthRouter := sub.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
	//	return r.Header.Get("Authorization") == ""
	//}).Subrouter()
	
	// instantiate a new controller which is supposed to serve our routes
	//controller := newController(cfg)
	
	// apply middleware
	var kcsvurl = cfg.KeyCloakURL
	var kcappurl = cfg.FrontendAppAddress
	keycloak.InitKeyCloak(kcsvurl,kcappurl,cfg)
	sub.Use(keycloak.AuthMiddleware)
	
	// map url routes to controller's methods
	/*
	sub.Path("/login").Methods("GET").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		//controller.login(writer, request)
		w.Write([]byte("Testing...."))
		//r.HTML(w, http.StatusOK, "login.html", nil)
	})
	*/

	//log.Println("Public Files for Login ", statikFS)
	sub.Path("/").Methods("GET").HandlerFunc(FormHandler(store, config))
	//sub.Path("/login").Methods("POST").HandlerFunc(LoginHandler(store,cfg))
	//sub.Path("/login").Methods("GET").HandlerFunc(keycloak.HandleLogin)
	sub.Path("/loginCallback").Methods("GET").HandlerFunc(keycloak.HandleLoginCallback)
	sub.Path("/logout").Methods("GET").HandlerFunc(keycloak.Logout)
	//sub.Path("/login").Methods("GET").HandlerFunc(LoginHandler(store,cfg))
	sub.Path("/add").Methods("POST").HandlerFunc(AddHandler(store, config))
	sub.Path("/remove").Methods("POST").HandlerFunc(RemoveHandler(store, config))
	sub.Path("/block").Methods("POST").HandlerFunc(BlockHandler(store, config))
	sub.PathPrefix("/public/").Handler(
		http.StripPrefix(config.Prefix+"/public/", http.FileServer(statikFS)))

	frontend := &Frontend{
		server: &http.Server{
			Handler:      CSRF(router),
			Addr:         address,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}

	frontend.done.Add(1)
	go func() {
		defer frontend.done.Done()
		log.Println("Frontend Listening on", frontend.server.Addr)
		if err := frontend.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Println(err)
		}
	}()
	return frontend
}

func (frontend *Frontend) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	if err := frontend.server.Shutdown(ctx); err != nil {
		log.Println("frontend shutdown:", err)
	}
	frontend.done.Wait()
}

type API struct {
	server *http.Server
	done   sync.WaitGroup
}

func NewAPI(address string, config ServerConfig, store *store.Store) *API {
	router := mux.NewRouter()
	router.Path("/publish").Methods("POST").HandlerFunc(PublishHandler(store))
	router.Path("/unpublish").Methods("POST").HandlerFunc(UnpublishHandler(store))

	api := &API{
		server: &http.Server{
			Handler:      router,
			Addr:         address,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}

	api.done.Add(1)
	go func() {
		defer api.done.Done()
		log.Println("API Listening on", api.server.Addr)
		if err := api.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	return api
}

func (api *API) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	if err := api.server.Shutdown(ctx); err != nil {
		log.Println("api shutdown:", err)
	}
	api.done.Wait()
}
