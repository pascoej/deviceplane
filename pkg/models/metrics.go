package models

import "time"

// ServiceMetricConfig is used for metrics scraping
type ServiceMetricConfig struct {
	Port uint   `json:"port" yaml:"port"`
	Path string `json:"path" yaml:"path"`
}

// The following are used for metrics forwarding

type MetricTargetConfig struct {
	ID        string           `json:"id" yaml:"id"`
	CreatedAt time.Time        `json:"createdAt" yaml:"createdAt"`
	ProjectID string           `json:"projectId" yaml:"projectId"`
	Type      MetricTargetType `json:"type" yaml:"type"`
	Configs   []MetricConfig   `json:"configs" yaml:"configs"`
}

type MetricTargetType string

const (
	MetricServiceTargetType MetricTargetType = "service"
	MetricHostTargetType    MetricTargetType = "host"
	MetricStateTargetType   MetricTargetType = "state"
)

type MetricConfig struct {
	Params  *ServiceMetricParams `json:"params,omitempty" json:"yaml,omitempty"`
	Metrics []Metric             `json:"metrics" yaml:"metrics"`
}

type ServiceMetricParams struct {
	ApplicationID string `json:"applicationId" yaml:"applicationId"`
	Service       string `json:"service" yaml:"service"`
}

type Metric struct {
	Metric string   `json:"metric" yaml:"metric"`
	Labels []string `json:"labels" yaml:"labels"`
	Tags   []string `json:"tags" yaml:"tags"`
}
