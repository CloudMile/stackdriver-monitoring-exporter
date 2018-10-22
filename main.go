package main

// https://github.com/googleapis/google-api-go-client/tree/master/monitoring/v3
// https://godoc.org/golang.org/x/oauth2/google#DefaultClient

import (
	"stackdriver-monitoring-exporter/pkg/gcp/stackdriver"
	"stackdriver-monitoring-exporter/pkg/metric_exporter"
	"stackdriver-monitoring-exporter/pkg/utils"
)

var monitoringMetrics = []string{
	"compute.googleapis.com/instance/cpu/usage_time",
	"compute.googleapis.com/instance/network/sent_bytes_count",
	"compute.googleapis.com/instance/network/received_bytes_count",
}

////////////////////////////////////////////////////////////////
// Main Function

func main() {
	var c utils.Conf
	c.LoadConfig()

	client := stackdriver.MonitoringClient{}
	exporter := metric_exporter.FileExporter{}

	client.SetTimezone(8)

	for prjIdx := range c.Projects {
		projectID := c.Projects[prjIdx].ProjectID

		for mIdx := range monitoringMetrics {
			metric := monitoringMetrics[mIdx]

			instanceNames := client.GetInstanceNames(projectID, metric)

			for instIdx := range instanceNames {
				instanceName := instanceNames[instIdx]

				points := client.RetrieveMetricPoints(projectID, metric, instanceName)

				exporter.Export(client.StartTime.In(client.Location()), projectID, metric, instanceName, points)
			}
		}
	}
}
