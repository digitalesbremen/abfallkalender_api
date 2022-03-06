package main

import (
	api "abfallkalender_api/src/backend"
	_ "embed"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
)

//go:embed dist/kalender.js
var webComponentJS string

func main() {
	log.Println("Hello Bremer Abfallkalender API!!!")

	router := api.NewRouter(webComponentJS)

	port := os.Getenv("PORT") // Heroku provides the port to bind to
	if port == "" {
		port = "8080"
	}

	log.Printf("Port is set to %s\n", port)

	log.Fatal(http.ListenAndServe(":"+port,
		handlers.CompressHandler(
			handlers.CORS(
				handlers.AllowedOrigins([]string{"*"}))(router))))
}
