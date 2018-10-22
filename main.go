package main

// https://github.com/googleapis/google-api-go-client/tree/master/monitoring/v3
// https://godoc.org/golang.org/x/oauth2/google#DefaultClient

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"stackdriver-monitoring-exporter/pkg/gcp/stackdriver"
	. "stackdriver-monitoring-exporter/pkg/utils"
)

const dir = "metrics"

var monitoringMetrics = []string{
	"compute.googleapis.com/instance/cpu/usage_time",
	"compute.googleapis.com/instance/network/sent_bytes_count",
	"compute.googleapis.com/instance/network/received_bytes_count",
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


func export(dateTime time.Time, projectID, metric, instanceName string, metricPoints []string) {
	folder := fmt.Sprintf("%s/%s/%d/%2d/%2d/%s", dir, projectID, dateTime.Year(), dateTime.Month(), dateTime.Day(), instanceName)
	os.MkdirAll(folder, os.ModePerm)

	title := strings.Replace(metric, "compute.googleapis.com/instance/", "", -1)
	title = strings.Replace(title, "/", "_", -1)

	output := fmt.Sprintf("%s/%s[%s][%s].csv", folder, dateTime.Format("2006-01-02T15:04:05"), instanceName, title)
	saveTimeSeriesToCSV(output, metricPoints)
}

////////////////////////////////////////////////////////////////
// Main Function

func main() {
	var c Conf
	c.LoadConfig()

	client := stackdriver.MonitoringClient{}

	client.SetTimezone(8)

	for prjIdx := range c.Projects {
		projectID := c.Projects[prjIdx].ProjectID

		for mIdx := range monitoringMetrics {
			metric := monitoringMetrics[mIdx]

			instanceNames := client.GetInstanceNames(projectID, metric)

			for instIdx := range instanceNames {
				instanceName := instanceNames[instIdx]

				points := client.RetrieveMetricPoints(projectID, metric, instanceName)

				export(client.StartTime.In(client.Location()), projectID, metric, instanceName, points)
			}
		}
	}
}
