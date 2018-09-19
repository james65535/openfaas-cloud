package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/openfaas/openfaas-cloud/auth/handlers"
)

func main() {
	var clientID string
	var clientSecret string
	var externalRedirectDomain string
	var cookieRootDomain string

	if val, exists := os.LookupEnv("client_id"); exists {
		clientID = val
	}

	if val, exists := os.LookupEnv("client_secret"); exists {
		clientSecret = val
	}

	if val, exists := os.LookupEnv("external_redirect_domain"); exists {
		externalRedirectDomain = val
	}

	if val, exists := os.LookupEnv("cookie_root_domain"); exists {
		cookieRootDomain = val
	}

	config := &handlers.Config{
		ClientID:               clientID,
		ClientSecret:           clientSecret,
		CookieExpiresIn:        time.Hour * 48,
		CookieRootDomain:       cookieRootDomain,
		ExternalRedirectDomain: externalRedirectDomain,
		Scope: "read:org,read:user,user:email",
	}

	protected := []string{
		"/function/system-dashboard",
		"/function/system-list-functions",
	}

	fs := http.FileServer(http.Dir("static"))

	router := http.NewServeMux()
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	router.HandleFunc("/", handlers.MakeHomepageHandler(config))

	router.HandleFunc("/q/", handlers.MakeQueryHandler(config, protected))
	router.HandleFunc("/login/", handlers.MakeLoginHandler(config))
	router.HandleFunc("/oauth2/", handlers.MakeOAuth2Handler(config))
	router.HandleFunc("/healthz/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK."))
	})

	timeout := time.Second * 10
	port := 8080
	if v, exists := os.LookupEnv("port"); exists {
		val, _ := strconv.Atoi(v)
		port = val
	}

	log.Printf("Using port: %d\n", port)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        router,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}