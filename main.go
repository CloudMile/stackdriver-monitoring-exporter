package main

// https://github.com/googleapis/google-api-go-client/tree/master/monitoring/v3
// https://godoc.org/golang.org/x/oauth2/google#DefaultClient

import (
	"fmt"
	"log"
	"os"
	"strings"

	"stackdriver-monitoring-exporter/pkg/gcp/stackdriver"
	. "stackdriver-monitoring-exporter/pkg/utils"
)

const dir = "metrics"

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

////////////////////////////////////////////////////////////////
// Main Function

func main() {
	var c Conf
	c.LoadConfig()

	client := stackdriver.MonitoringClient{}

	for i := range c.Projects {
		project := c.Projects[i]
		for j := range project.Metrics {
			metric := project.Metrics[j]
			metric.LoadConfig()

			points := client.RetrieveMetricPoints(project.ProjectID, &metric)

			dateTime := metric.StartTime.In(metric.Location()).Format("2006-01-02T15:04:05")
			output := fmt.Sprintf("%s/%s[%s][%s].csv", dir, dateTime, metric.Title, metric.Unit)
			saveTimeSeriesToCSV(output, points)
		}
	}
}
