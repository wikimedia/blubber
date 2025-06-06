package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed dist
var embedFS embed.FS

func main() {
	distFS, err := fs.Sub(embedFS, "dist")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.FS(distFS)))

	port := ":8080"
	log.Printf("Listening at http://localhost%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
