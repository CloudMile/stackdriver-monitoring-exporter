package utils

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

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
