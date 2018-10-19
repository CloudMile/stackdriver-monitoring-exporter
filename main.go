package main

// https://github.com/googleapis/google-api-go-client/tree/master/monitoring/v3
// https://godoc.org/golang.org/x/oauth2/google#DefaultClient

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/monitoring/v3"
	. "stackdriver-monitoring-exporter/pkg/utils"
)

const dir = "metrics"

func getCred(ctx context.Context) (cred *google.Credentials) {
	cred, err := google.FindDefaultCredentials(ctx, monitoring.MonitoringReadScope)
	if err != nil {
		log.Fatal("%v", err)
	}
	log.Printf("Project ID: %s", cred.ProjectID)

	return
}

func newClient(ctx context.Context, cred *google.Credentials) (client *http.Client) {
	conf, err := google.JWTConfigFromJSON(cred.JSON, monitoring.MonitoringReadScope)
	if err != nil {
		log.Fatal(err)
	}

	client = conf.Client(ctx)

	return
}

func saveTimeSeriesToCSV(filename string, points []*monitoring.Point) {
	log.Printf("Points len: %d", len(points))

	n := len(points)
	metrics := make([]string, n+1)
	metrics[0] = "timestamp,datetime,value"
	var idx int
	for i := range points {
		idx = n - i - 1
		t, _ := time.Parse("2006-01-02T15:04:05Z", points[idx].Interval.StartTime)
		t = t.Add(time.Hour * 8)
		metrics[i+1] = fmt.Sprintf("%d,%s,%g", t.Unix(), t.Format("2006-01-02 15:04:05"), *(points[idx].Value.DoubleValue))
	}

	saveToFile(filename, strings.Join(metrics, "\n"))
}

func saveToFile(filename, content string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	fmt.Fprintf(file, content)
}

func retrieveMetricPoints(projectID string, cm *ConfMetric) (points []*monitoring.Point) {
	ctx := context.Background()
	cred := getCred(ctx)
	client := newClient(ctx, cred)

	svc, err := monitoring.New(client)
	if err != nil {
		log.Fatal("%v", err)
	}

	project := "projects/" + projectID

	projectsTimeSeriesListCall := svc.Projects.TimeSeries.List(project)
	projectsTimeSeriesListCall.Filter(cm.Filters())
	projectsTimeSeriesListCall.IntervalStartTime(cm.IntervalStartTime())
	projectsTimeSeriesListCall.IntervalEndTime(cm.IntervalEndTime())
	projectsTimeSeriesListCall.AggregationPerSeriesAligner(cm.AggregationPerSeriesAligner)
	projectsTimeSeriesListCall.AggregationAlignmentPeriod(cm.AggregationAlignmentPeriod)

	listResp, err := projectsTimeSeriesListCall.Do()
	if err != nil {
		log.Fatal("%v", err)
	}

	// Only get the first timeseries
	timeSeries := listResp.TimeSeries[0]
	points = timeSeries.Points

	return
}

////////////////////////////////////////////////////////////////
// Main Function

func main() {
	var c Conf
	c.LoadConfig()

	for i := range c.Projects {
		project := c.Projects[i]
		for j := range project.Metrics {
			metric := project.Metrics[j]
			metric.LoadConfig()

			points := retrieveMetricPoints(project.ProjectID, &metric)

			dateTime := metric.StartTime.In(metric.Location()).Format("2006-01-02T15:04:05")
			output := fmt.Sprintf("%s/%s[%s][%s].csv", dir, dateTime, metric.Title, metric.Unit)
			saveTimeSeriesToCSV(output, points)
		}
	}
}
