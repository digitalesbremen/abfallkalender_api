package main

import (
	api "abfallkalender_api/src/backend"
	_ "embed"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
)

// //go:embed dist/kalender.js
var kalenderJS string

// //go:embed dist/kalender.js.map
var kalenderJSMap string

func main() {
	log.Println("Hello Bremer Abfallkalender API!!!")

	router := api.NewRouter(kalenderJS, kalenderJSMap)

	// TODO use os.lookup env
	port, portSet := os.LookupEnv("PORT")
	if !portSet {
		port = "8080"
	}

	log.Printf("Port is set to %s\n", port)

	// TODO use go routine / non blocking
	log.Fatal(http.ListenAndServe(":"+port,
		handlers.CompressHandler(
			handlers.CORS(
				handlers.AllowedOrigins([]string{"*"}))(router))))
}
