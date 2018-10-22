package service

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

type ExportService struct {
	conf     utils.Conf
	Exporter metric_exporter.MetricExporter
}

func (es ExportService) Do() {
	var c utils.Conf
	c.LoadConfig()

	client := stackdriver.MonitoringClient{}

	client.SetTimezone(c.Timezone)

	for prjIdx := range c.Projects {
		projectID := c.Projects[prjIdx].ProjectID

		for mIdx := range monitoringMetrics {
			metric := monitoringMetrics[mIdx]

			instanceNames := client.GetInstanceNames(projectID, metric)

			for instIdx := range instanceNames {
				instanceName := instanceNames[instIdx]

				points := client.RetrieveMetricPoints(projectID, metric, instanceName)

				es.Exporter.Export(client.StartTime.In(client.Location()), projectID, metric, instanceName, points)
			}
		}
	}
}
