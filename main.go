package main

import (
	"abfallkalender_api/src/backend/client"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

//go:embed dist/kalender.js
var content string

const (
	BaseURL            = "https://web.c-trace.de"
	InitialContextPath = "/bremenabfallkalender/Abfallkalender"
)

func main() {
	log.Println("Hello Bremer Abfallkalender API!!!")

	router := mux.NewRouter()

	router.HandleFunc("/component", serveWebComponent).Methods("GET")
	router.HandleFunc("/streets", serveStreets).Methods("GET")
	router.HandleFunc("/", serveOpenApiSpecification).Methods("GET")

	http.Handle("/", router)

	port := os.Getenv("PORT") // Heroku provides the port to bind to
	if port == "" {
		port = "8080"
	}

	log.Printf("Port is set to %s\n", port)

	url, _ := client.NewClient(BaseURL).GetRedirectUrl(InitialContextPath)
	log.Printf("Base URL is %s", BaseURL)
	log.Printf("Initial context path is %s", InitialContextPath)
	log.Printf("Redirect URL is %s", url)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveWebComponent(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = fmt.Fprint(w, content)
}

func serveOpenApiSpecification(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/x-yaml; charset=UTF-8")
	dat, _ := ioutil.ReadFile("open-api-3.yaml")
	_, _ = w.Write(dat)
}

func serveStreets(w http.ResponseWriter, _ *http.Request) {
	abfallkalenderClient := client.NewClient(BaseURL)
	// TODO handle error
	redirectUrl, _ := abfallkalenderClient.GetRedirectUrl(InitialContextPath)
	// TODO handle error
	streets, _ := abfallkalenderClient.GetStreets(redirectUrl)
	// TODO handle error
	dto, _ := json.Marshal(streets.Streets)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(dto)
}
