package service

import (
	"context"
	"google.golang.org/appengine/taskqueue"
	"log"

	"stackdriver-monitoring-exporter/pkg/gcp/stackdriver"
	"stackdriver-monitoring-exporter/pkg/metric_exporter"
	"stackdriver-monitoring-exporter/pkg/utils"
)

var monitoringMetrics = []string{
	"compute.googleapis.com/instance/cpu/usage_time",
	"compute.googleapis.com/instance/network/sent_bytes_count",
	"compute.googleapis.com/instance/network/received_bytes_count",
}

// one instance may have many disks
var monitoringDiskMetrics = []string{
	"compute.googleapis.com/instance/disk/write_ops_count",
	"compute.googleapis.com/instance/disk/read_ops_count",
}

type ExportService struct {
	conf   utils.Conf
	client stackdriver.MonitoringClient
}

func NewExportService() ExportService {
	var es = ExportService{}
	return es.init()
}

func (es ExportService) newMetricExporter() metric_exporter.MetricExporter {
	switch es.conf.ExporterClass {
	case "GCSExporter":
		return metric_exporter.NewGCSExporter(es.conf)
	default:
		return metric_exporter.NewFileExporter(es.conf)
	}
}

func (es ExportService) init() ExportService {
	es.conf.LoadConfig()

	es.client = stackdriver.MonitoringClient{}
	es.client.SetTimezone(es.conf.Timezone)

	return es
}

func (es ExportService) Do(ctx context.Context) {
	es.client.SetContext(ctx)

	for prjIdx := range es.conf.Projects {
		projectID := es.conf.Projects[prjIdx].ProjectID

		// Common instance metrics
		es.exportInstanceCommonMetrics(ctx, projectID)

		// Disk metrics
		es.exportInstanceDiskMetrics(ctx, projectID)
	}
}

func (es ExportService) exportInstanceCommonMetrics(ctx context.Context, projectID string) {
	for mIdx := range monitoringMetrics {
		metric := monitoringMetrics[mIdx]

		log.Printf("es.client.GetInstanceNames")
		instanceNames := es.client.GetInstanceNames(projectID, metric)

		for instIdx := range instanceNames {
			instanceName := instanceNames[instIdx]

			filter := stackdriver.MakeInstanceFilter(metric, instanceName)
			// es.Export(projectID, metric, filter, instanceName)

			t := taskqueue.NewPOSTTask(
				"/export",
				map[string][]string{
					"projectID":    {projectID},
					"metric":       {metric},
					"filter":       {filter},
					"instanceName": {instanceName},
				},
			)
			if _, err := taskqueue.Add(ctx, t, ""); err != nil {
				log.Fatal(err.Error())
			}
		}
	}
}

func (es ExportService) exportInstanceDiskMetrics(ctx context.Context, projectID string) {
	for mdIdx := range monitoringDiskMetrics {
		metric := monitoringDiskMetrics[mdIdx]

		instanceAndDiskMaps := es.client.GetInstanceAndDiskMaps(projectID, metric)

		for mapIdx := range instanceAndDiskMaps {
			m := instanceAndDiskMaps[mapIdx]
			instanceName := m[stackdriver.InstanceNameKey]
			deviceName := m[stackdriver.DeviceNameKey]

			filter := stackdriver.MakeDiskFilter(metric, instanceName, deviceName)
			// es.Export(projectID, metric, filter, instanceName, "disk", deviceName)

			t := taskqueue.NewPOSTTask(
				"/export",
				map[string][]string{
					"projectID":    {projectID},
					"metric":       {metric},
					"filter":       {filter},
					"instanceName": {instanceName},
					"attendNames":  {"disk", deviceName},
				},
			)
			if _, err := taskqueue.Add(ctx, t, ""); err != nil {
				log.Fatal(err.Error())
			}
		}
	}
}

func (es ExportService) Export(projectID, metric, filter, instanceName string, attendNames ...string) {
	points := es.client.RetrieveMetricPoints(projectID, metric, filter)

	metricExporter := es.newMetricExporter()
	metricExporter.Export(es.client.StartTime.In(es.client.Location()), projectID, metric, instanceName, points, attendNames...)
}
