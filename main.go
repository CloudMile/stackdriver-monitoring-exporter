package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"stackdriver-monitoring-exporter/pkg/service"
)

func main() {
	http.HandleFunc("/", indexHandler)

	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	// [END setting_port]
}

// Index
func indexHandler(w http.ResponseWriter, r *http.Request) {
	exportService := service.NewExportService()
	exportService.Do()

	fmt.Fprint(w, "Done")
}
