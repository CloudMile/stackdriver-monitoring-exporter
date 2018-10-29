package main

import (
	"fmt"
	"google.golang.org/appengine"
	"log"
	"net/http"
	"os"
	"stackdriver-monitoring-exporter/pkg/service"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/export", exportMetricPointsHandler)

	appengine.Main()
}

// Index
func indexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	exportService := service.NewExportService()
	exportService.Do(ctx)

	fmt.Fprint(w, "Done")
}

// Export Metric Points to CSV
func exportMetricPointsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Done")
}
