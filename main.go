package main

import (
	"stackdriver-monitoring-exporter/pkg/service"
)

func main() {
	exportService := service.ExportService{}
	exportService.Do()
}
