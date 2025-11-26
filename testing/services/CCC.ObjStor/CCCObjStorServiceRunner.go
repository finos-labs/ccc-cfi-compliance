package objstor

import (
	"context"
	"fmt"
	"log"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/factory"
	objstorage "github.com/finos-labs/ccc-cfi-compliance/testing/api/object-storage"
	"github.com/finos-labs/ccc-cfi-compliance/testing/inspection"
	"github.com/finos-labs/ccc-cfi-compliance/testing/services"
)

// CCCObjStorServiceRunner implements the ServiceRunner interface for object storage compliance tests
type CCCObjStorServiceRunner struct {
	*services.AbstractServiceRunner
}

// NewCCCObjStorServiceRunner creates a new object storage service runner
func NewCCCObjStorServiceRunner(config services.RunConfig) *CCCObjStorServiceRunner {
	runner := &CCCObjStorServiceRunner{}
	runner.AbstractServiceRunner = services.NewAbstractServiceRunner(
		"CCC.ObjStor",
		runner.GetTestResources,
		config,
		"features", // Features path relative to this service's directory
	)
	return runner
}

// GetTestResources discovers object storage buckets/containers for the specified provider
func (r *CCCObjStorServiceRunner) GetTestResources(ctx context.Context, provider string, cloudFactory factory.Factory) ([]inspection.TestParams, error) {
	// Get object storage service
	service, err := cloudFactory.GetServiceAPI("object-storage")
	if err != nil {
		return nil, fmt.Errorf("failed to create object storage service: %w", err)
	}

	storageService, ok := service.(objstorage.Service)
	if !ok {
		return nil, fmt.Errorf("service does not implement object storage interface")
	}

	// List all buckets
	buckets, err := storageService.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	// Enrich buckets with region information
	for i := range buckets {
		if buckets[i].Region == "" {
			region, err := storageService.GetBucketRegion(buckets[i].ID)
			if err != nil {
				log.Printf("⚠️  Warning: Failed to get region for bucket %s: %v", buckets[i].ID, err)
			} else {
				buckets[i].Region = region
			}
		}
	}

	// Convert buckets to TestParams
	resources := make([]inspection.TestParams, 0, len(buckets))
	for _, bucket := range buckets {
		resources = append(resources, inspection.TestParams{
			Provider:            provider,
			ResourceName:        bucket.Name,
			UID:                 bucket.ID,
			ProviderServiceType: "object-storage",
			CatalogType:         "CCC.ObjStor",
			Region:              bucket.Region,
		})
	}

	return resources, nil
}
