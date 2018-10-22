package main

import (
	"stackdriver-monitoring-exporter/pkg/metric_exporter"
	"stackdriver-monitoring-exporter/pkg/service"
)

func main() {
	exportService := service.ExportService{}
	exportService.Exporter = metric_exporter.FileExporter{}

	exportService.Do()
}
