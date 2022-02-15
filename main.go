package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	log.Println("Hello Bremer Abfallkalender API!")

	router := mux.NewRouter()

	router.HandleFunc("/", helloWorld).Methods("GET")
	http.Handle("/", router)

	_ = http.ListenAndServe(":8080", router)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hallo Bremer Abfallkalender API!")
}
