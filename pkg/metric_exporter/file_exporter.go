package metric_exporter

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"stackdriver-monitoring-exporter/pkg/gcp/stackdriver"
)

const dir = "metrics"

type FileExporter struct {
}

func saveTimeSeriesToCSV(filename string, metricPoints []string) {
	log.Printf("Points len: %d", len(metricPoints))

	saveToFile(filename, stackdriver.PointCSVHeader, strings.Join(metricPoints, "\n"))
}

func saveToFile(filename, header, content string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "%s\n", header)
	fmt.Fprintf(file, content)
}

func (f FileExporter) Export(dateTime time.Time, projectID, metric, instanceName string, metricPoints []string) {
	folder := fmt.Sprintf("%s/%s/%d/%2d/%2d/%s", dir, projectID, dateTime.Year(), dateTime.Month(), dateTime.Day(), instanceName)
	os.MkdirAll(folder, os.ModePerm)

	title := strings.Replace(metric, "compute.googleapis.com/instance/", "", -1)
	title = strings.Replace(title, "/", "_", -1)

	output := fmt.Sprintf("%s/%s[%s][%s].csv", folder, dateTime.Format("2006-01-02T15:04:05"), instanceName, title)
	saveTimeSeriesToCSV(output, metricPoints)
}
