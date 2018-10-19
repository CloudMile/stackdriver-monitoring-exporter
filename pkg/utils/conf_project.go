package utils

type ConfProject struct {
	ProjectID string       `yaml:"projectID"`
	Metrics   []ConfMetric `yaml:"metrics"`
}
