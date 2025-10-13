package inspection

import (
	"embed"
	"encoding/csv"
	"fmt"
	"strings"
	"sync"
)

//go:embed *.csv
var csvFiles embed.FS

// ServiceMapping represents a mapping between provider service type and catalog type
type ServiceMapping struct {
	ProviderServiceType string
	CatalogType         string
	Description         string
}

// ServiceMappingRegistry holds all service mappings
type ServiceMappingRegistry struct {
	mappings map[string]map[string]ServiceMapping // provider -> service_type -> mapping
	mu       sync.RWMutex
}

var (
	registry     *ServiceMappingRegistry
	registryOnce sync.Once
)

// GetRegistry returns the singleton service mapping registry
func GetRegistry() *ServiceMappingRegistry {
	registryOnce.Do(func() {
		registry = &ServiceMappingRegistry{
			mappings: make(map[string]map[string]ServiceMapping),
		}
		// Load mappings on first access
		if err := registry.loadMappings(); err != nil {
			// Log error but don't panic - allow runtime to continue
			fmt.Printf("Warning: failed to load service mappings: %v\n", err)
		}
	})
	return registry
}

// loadMappings loads all CSV mapping files
func (r *ServiceMappingRegistry) loadMappings() error {
	providers := []string{"aws", "azure", "gcp"}

	for _, provider := range providers {
		filename := fmt.Sprintf("%s-services.csv", provider)
		if err := r.loadProviderMappings(provider, filename); err != nil {
			return fmt.Errorf("failed to load %s: %w", filename, err)
		}
	}

	return nil
}

// loadProviderMappings loads mappings from a CSV file for a specific provider
func (r *ServiceMappingRegistry) loadProviderMappings(provider, filename string) error {
	data, err := csvFiles.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Skip header row
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mappings[provider] == nil {
		r.mappings[provider] = make(map[string]ServiceMapping)
	}

	for i, record := range records[1:] {
		if len(record) < 3 {
			return fmt.Errorf("invalid CSV format at row %d: expected 3 columns, got %d", i+2, len(record))
		}

		mapping := ServiceMapping{
			ProviderServiceType: record[0],
			CatalogType:         record[1],
			Description:         record[2],
		}

		r.mappings[provider][mapping.ProviderServiceType] = mapping
	}

	return nil
}

// LookupCatalogType returns the catalog type for a given provider and service type
func (r *ServiceMappingRegistry) LookupCatalogType(provider, providerServiceType string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerMappings, ok := r.mappings[provider]
	if !ok {
		return "", false
	}

	mapping, ok := providerMappings[providerServiceType]
	if !ok {
		return "", false
	}

	return mapping.CatalogType, true
}

// GetMapping returns the full mapping for a given provider and service type
func (r *ServiceMappingRegistry) GetMapping(provider, providerServiceType string) (ServiceMapping, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerMappings, ok := r.mappings[provider]
	if !ok {
		return ServiceMapping{}, false
	}

	mapping, ok := providerMappings[providerServiceType]
	return mapping, ok
}

// GetAllMappings returns all mappings for a given provider
func (r *ServiceMappingRegistry) GetAllMappings(provider string) []ServiceMapping {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerMappings, ok := r.mappings[provider]
	if !ok {
		return nil
	}

	result := make([]ServiceMapping, 0, len(providerMappings))
	for _, mapping := range providerMappings {
		result = append(result, mapping)
	}

	return result
}

// LookupCatalogType is a convenience function that uses the default registry
func LookupCatalogType(provider, providerServiceType string) (string, bool) {
	return GetRegistry().LookupCatalogType(provider, providerServiceType)
}
