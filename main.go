package main

import (
	_ "embed"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

//go:embed dist/kalender.js
var content string

func main() {
	log.Println("Hello Bremer Abfallkalender API!!!")

	router := mux.NewRouter()

	router.HandleFunc("/component", serveJs).Methods("GET")
	router.HandleFunc("/", helloWorld).Methods("GET")

	http.Handle("/", router)

	port := os.Getenv("PORT") // Heroku provides the port to bind to
	if port == "" {
		port = "8080"
	}

	log.Printf("Port is set to %s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveJs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = fmt.Fprint(w, content)
}
func helloWorld(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hallo Bremer Abfallkalender API!")
}
