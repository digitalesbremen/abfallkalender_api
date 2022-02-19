package main

import (
	_ "embed"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

//go:embed dist/kalender.js
var content string

func main() {
	log.Println("Hello Bremer Abfallkalender API!!!")

	router := mux.NewRouter()

	router.HandleFunc("/component", serveWebComponent).Methods("GET")
	router.HandleFunc("/hello", helloWorld).Methods("GET")
	router.HandleFunc("/", serveOpenApiSpecification).Methods("GET")

	http.Handle("/", router)

	port := os.Getenv("PORT") // Heroku provides the port to bind to
	if port == "" {
		port = "8080"
	}

	log.Printf("Port is set to %s\n", port)

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

func helloWorld(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hallo Bremer Abfallkalender API!")
}
