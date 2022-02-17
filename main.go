package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("Hello Bremer Abfallkalender API!!!")

	router := mux.NewRouter()

	router.HandleFunc("/", helloWorld).Methods("GET")
	http.Handle("/", router)

	port := os.Getenv("PORT") // Heroku provides the port to bind to
	if port == "" {
		port = "8080"
	}

	log.Printf("Port is set to %s\n", port)

	_ = http.ListenAndServe(":"+port, router)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hallo Bremer Abfallkalender API!")
}
