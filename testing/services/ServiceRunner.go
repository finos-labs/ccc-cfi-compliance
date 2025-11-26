package services

import "time"

// RunConfig is the configuration for running compliance tests
type RunConfig struct {
	Provider       string
	OutputDir      string
	Timeout        time.Duration
	ResourceFilter string
}

// ServiceRunner is the interface for running compliance tests for a specific service
type ServiceRunner interface {
	Run() int

	GetConfig() RunConfig
}
