package stackdriver

import (
	"log"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/monitoring/v3"
	. "stackdriver-monitoring-exporter/pkg/utils"
)

type MonitoringClient struct {
}

func (c *MonitoringClient) getCred(ctx context.Context) (cred *google.Credentials) {
	cred, err := google.FindDefaultCredentials(ctx, monitoring.MonitoringReadScope)
	if err != nil {
		log.Fatal("%v", err)
	}
	log.Printf("Project ID: %s", cred.ProjectID)

	return
}

func (c *MonitoringClient) newClient(ctx context.Context, cred *google.Credentials) (client *http.Client) {
	conf, err := google.JWTConfigFromJSON(cred.JSON, monitoring.MonitoringReadScope)
	if err != nil {
		log.Fatal(err)
	}

	client = conf.Client(ctx)

	return
}

func (c *MonitoringClient) RetrieveMetricPoints(projectID string, cm *ConfMetric) (points []*monitoring.Point) {
	ctx := context.Background()
	cred := c.getCred(ctx)
	client := c.newClient(ctx, cred)

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
