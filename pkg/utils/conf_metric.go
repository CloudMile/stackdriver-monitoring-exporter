package utils

import (
	"fmt"
	"strings"
	"time"
)

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
