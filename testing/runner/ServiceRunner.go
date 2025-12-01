package main

import (
	"time"

	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

// RunConfig is the configuration for running compliance tests
type RunConfig struct {
	ServiceName    string // e.g., "object-storage", "iam"
	CloudParams    environment.CloudParams
	OutputDir      string
	Timeout        time.Duration
	ResourceFilter string
	Tag            string // Optional tag filter to override automatic catalog type filtering
}

// ServiceRunner is the interface for running compliance tests for a specific service
type ServiceRunner interface {
	Run() int

	GetConfig() RunConfig
}
