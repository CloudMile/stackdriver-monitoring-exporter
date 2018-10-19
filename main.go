package main

// https://github.com/googleapis/google-api-go-client/tree/master/monitoring/v3
// https://godoc.org/golang.org/x/oauth2/google#DefaultClient

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/monitoring/v3"

	"gopkg.in/yaml.v2"
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
// Config File

type ConfMetric struct {
	Title                       string   `yaml:"title"`
	MetricType                  string   `yaml:"metricType"`
	AndFilters                  []string `yaml:"andFilters,flow"`
	IntervalType                string   `yaml:"intervalType"` // Month, Day, Hour
	TimeZone                    int64    `yaml:"timezone"`
	AggregationPerSeriesAligner string   `yaml:"aggregationPerSeriesAligner"`
	AggregationAlignmentPeriod  string   `yaml:"aggregationAlignmentPeriod"`
	Unit                        string   `yaml:"unit"`
	StartTime                   time.Time
	EndTime                     time.Time
}

func (cm *ConfMetric) LoadConfig() *ConfMetric {
	local := cm.Location()
	now := time.Now().In(local)

	intervalType := cm.IntervalType
	switch intervalType {
	case "month":
		cm.EndTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, local).UTC()
		cm.StartTime = cm.EndTime.AddDate(0, -1, 0)
	case "day":
		cm.EndTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, local).UTC()
		cm.StartTime = cm.EndTime.AddDate(0, 0, -1)
	case "hour":
		cm.EndTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, local).UTC()
		cm.StartTime = cm.EndTime.Add(-1 * time.Hour)
	}

	return cm
}

func (c *ConfMetric) Filters() string {
	len := len(c.AndFilters) + 1
	filters := make([]string, len)
	filters[0] = fmt.Sprintf(`metric.type="%s"`, c.MetricType)

	for idx := range c.AndFilters {
		filters[idx+1] = c.AndFilters[idx]
	}

	return strings.Join(filters, " AND ")
}

func (c *ConfMetric) IntervalStartTime() string {
	return c.StartTime.Format("2006-01-02T15:04:05.000000000Z")
}

func (c *ConfMetric) IntervalEndTime() string {
	return c.EndTime.Format("2006-01-02T15:04:05.000000000Z")
}

func (c *ConfMetric) Location() *time.Location {
	localSecondsEastOfUTC := int((time.Duration(c.TimeZone) * time.Hour).Seconds())
	return time.FixedZone("localtime", localSecondsEastOfUTC)
}

type ConfProject struct {
	ProjectID string       `yaml:"projectID"`
	Metrics   []ConfMetric `yaml:"metrics"`
}

type Conf struct {
	Projects []ConfProject `yaml:"projects"`
}

func (c *Conf) LoadConfig() *Conf {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	var project ConfProject
	var metric ConfMetric
	for i := range c.Projects {
		project = c.Projects[i]
		for j := range project.Metrics {
			metric = project.Metrics[j]
			metric.LoadConfig()
		}
	}

	return c
}

// func (c *Conf) Log() {
// 	log.Println("Load config...")
// 	log.Println("ProjectID                  : ", c.ProjectID)
// 	log.Println("Title                      : ", c.Title)
// 	log.Println("MetricType                 : ", c.MetricType)
// 	log.Println("AndFilters                 : ", c.AndFilters)
// 	log.Println("IntervalType               : ", c.IntervalType)
// 	log.Println("TimeZone                   : ", c.TimeZone)
// 	log.Println("AggregationPerSeriesAligner: ", c.AggregationPerSeriesAligner)
// 	log.Println("AggregationAlignmentPeriod : ", c.AggregationAlignmentPeriod)
// 	log.Println("intervalStartTime          : ", c.IntervalStartTime())
// 	log.Println("intervalEndTime            : ", c.IntervalEndTime())
// }

////////////////////////////////////////////////////////////////
// Main Function

func main() {
	var c Conf
	c.LoadConfig()
	// c.Log()

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
