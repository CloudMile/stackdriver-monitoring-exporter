package stackdriver

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/monitoring/v3"
)

const PointCSVHeader = "timestamp,datetime,value"
const AggregationAlignmentPeriod = "60s"
const AggregationPerSeriesAligner = "ALIGN_RATE"
const MinutesOneDay = 60 * 24

const InstanceNameKey = "instanceName"
const DeviceNameKey = "deviceName"

type MonitoringClient struct {
	TimeZone          int
	StartTime         time.Time
	EndTime           time.Time
	IntervalStartTime string
	IntervalEndTime   string
	client            *http.Client
}

func (c *MonitoringClient) SetTimezone(timezone int) {
	c.TimeZone = timezone

	local := c.Location()
	now := time.Now().In(local)

	c.EndTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, local).UTC()
	c.StartTime = c.EndTime.AddDate(0, 0, -1)

	c.IntervalEndTime = c.EndTime.Format("2006-01-02T15:04:05.000000000Z")
	c.IntervalStartTime = c.StartTime.Format("2006-01-02T15:04:05.000000000Z")

	log.Printf("%s", c.IntervalEndTime)
	log.Printf("%s", c.IntervalStartTime)
}

func (c *MonitoringClient) Location() *time.Location {
	localSecondsEastOfUTC := int((time.Duration(c.TimeZone) * time.Hour).Seconds())
	return time.FixedZone("localtime", localSecondsEastOfUTC)
}

func (c *MonitoringClient) getCred(ctx context.Context) (cred *google.Credentials) {
	cred, err := google.FindDefaultCredentials(ctx, monitoring.MonitoringReadScope)
	if err != nil {
		log.Fatal("%v", err)
	}
	log.Printf("Project ID: %s", cred.ProjectID)

	return
}

func (c *MonitoringClient) getClient() (client *http.Client) {
	if c.client == nil {
		ctx := context.Background()
		cred := c.getCred(ctx)
		c.client = c.newClient(ctx, cred)
	}

	client = c.client

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

func (c *MonitoringClient) pointsToMetricPoints(points []*monitoring.Point) (metricPoints []string) {
	metricPoints = make([]string, MinutesOneDay)

	pointTime := c.StartTime
	var pointIdx = len(points) - 1
	for metricIdx := range metricPoints {
		pointTime = pointTime.Add(time.Second * 60)

		t, _ := time.Parse("2006-01-02T15:04:05Z", points[pointIdx].Interval.StartTime)

		if pointTime.Equal(t) {
			t = t.Add(time.Hour * (time.Duration)(c.TimeZone))
			metricPoints[metricIdx] = fmt.Sprintf("%d,%s,%g", t.Unix(), t.Format("2006-01-02 15:04:05"), *(points[pointIdx].Value.DoubleValue))

			pointIdx = pointIdx - 1
		} else {
			t = pointTime.Add(time.Hour * (time.Duration)(c.TimeZone))
			metricPoints[metricIdx] = fmt.Sprintf("%d,%s,", t.Unix(), t.Format("2006-01-02 15:04:05"))
		}
	}

	return
}

func MakeInstanceFilter(metric, instanceName string) string {
	return fmt.Sprintf(`metric.type="%s" AND metric.labels.instance_name="%s"`, metric, instanceName)
}

func MakeDiskFilter(metric, instanceName, deviceName string) string {
	return fmt.Sprintf(`metric.type="%s" AND metric.labels.instance_name="%s" AND metric.labels.device_name="%s"`, metric, instanceName, deviceName)
}

func (c *MonitoringClient) RetrieveMetricPoints(projectID, metric, filter string) (metricPoints []string) {
	client := c.getClient()

	svc, err := monitoring.New(client)
	if err != nil {
		log.Fatal("%v", err)
	}

	project := "projects/" + projectID

	projectsTimeSeriesListCall := svc.Projects.TimeSeries.List(project)
	projectsTimeSeriesListCall.Filter(filter)
	projectsTimeSeriesListCall.IntervalStartTime(c.IntervalStartTime)
	projectsTimeSeriesListCall.IntervalEndTime(c.IntervalEndTime)
	projectsTimeSeriesListCall.AggregationPerSeriesAligner(AggregationPerSeriesAligner)
	projectsTimeSeriesListCall.AggregationAlignmentPeriod(AggregationAlignmentPeriod)

	listResp, err := projectsTimeSeriesListCall.Do()
	if err != nil {
		log.Fatal("%v", err)
	}

	// Only get the first timeseries
	timeSeries := listResp.TimeSeries[0]
	metricPoints = c.pointsToMetricPoints(timeSeries.Points)

	return
}

func (c *MonitoringClient) GetInstanceNames(projectID, metric string) (instanceNames []string) {
	client := c.getClient()

	svc, err := monitoring.New(client)
	if err != nil {
		log.Fatal("%v", err)
	}

	project := "projects/" + projectID

	projectsTimeSeriesListCall := svc.Projects.TimeSeries.List(project)
	projectsTimeSeriesListCall.View("HEADERS")
	projectsTimeSeriesListCall.Filter(`metric.type="` + metric + `"`)
	projectsTimeSeriesListCall.IntervalStartTime(c.IntervalStartTime)
	projectsTimeSeriesListCall.IntervalEndTime(c.IntervalEndTime)

	listResp, err := projectsTimeSeriesListCall.Do()
	if err != nil {
		log.Fatal("%v", err)
	}

	instanceNames = make([]string, len(listResp.TimeSeries))
	for i := range listResp.TimeSeries {
		instanceNames[i] = listResp.TimeSeries[i].Metric.Labels["instance_name"]
	}

	return
}

func (c *MonitoringClient) GetInstanceAndDiskMaps(projectID, diskMetric string) (instanceAndDiskMaps []map[string]string) {
	client := c.getClient()

	svc, err := monitoring.New(client)
	if err != nil {
		log.Fatal("%v", err)
	}

	project := "projects/" + projectID

	projectsTimeSeriesListCall := svc.Projects.TimeSeries.List(project)
	projectsTimeSeriesListCall.View("HEADERS")
	projectsTimeSeriesListCall.Filter(`metric.type="` + diskMetric + `"`)
	projectsTimeSeriesListCall.IntervalStartTime(c.IntervalStartTime)
	projectsTimeSeriesListCall.IntervalEndTime(c.IntervalEndTime)

	listResp, err := projectsTimeSeriesListCall.Do()
	if err != nil {
		log.Fatal("%v", err)
	}

	instanceAndDiskMaps = make([]map[string]string, len(listResp.TimeSeries))
	for i := range listResp.TimeSeries {
		m := make(map[string]string)
		m[InstanceNameKey] = listResp.TimeSeries[i].Metric.Labels["instance_name"]
		m[DeviceNameKey] = listResp.TimeSeries[i].Metric.Labels["device_name"]
		instanceAndDiskMaps[i] = m
	}

	return
}
