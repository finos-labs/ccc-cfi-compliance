package types

import "gopkg.in/yaml.v3"

// EnvironmentConfig is the top-level structure of types.yaml
type EnvironmentConfig struct {
	Instances []InstanceConfig `yaml:"instances"`
}

// InstanceConfig represents a named cloud environment instance
type InstanceConfig struct {
	ID         string                 `yaml:"id"`
	Type       string                 `yaml:"type"`
	Properties CloudParams            `yaml:"properties"`
	Services   []ServiceConfig        `yaml:"services"`
	Rules      map[string]interface{} `yaml:"rules"`
}

// ServiceConfig represents a service within an instance.
// The "type" key identifies the service; all other keys are service-specific properties.
type ServiceConfig struct {
	Type       string
	Properties map[string]interface{}
}

// CloudParams returns the instance's CloudParams with Provider set from the instance type.
func (ic InstanceConfig) CloudParams() CloudParams {
	params := ic.Properties
	params.Provider = ic.Type
	return params
}

// ServiceProperties returns the properties map for the named service type, or nil if not found.
func (ic InstanceConfig) ServiceProperties(serviceType string) map[string]interface{} {
	for _, svc := range ic.Services {
		if svc.Type == serviceType {
			return svc.Properties
		}
	}
	return nil
}

func (s *ServiceConfig) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if t, ok := raw["type"].(string); ok {
		s.Type = t
	}
	s.Properties = make(map[string]interface{})
	for k, v := range raw {
		if k != "type" {
			s.Properties[k] = v
		}
	}
	return nil
}
