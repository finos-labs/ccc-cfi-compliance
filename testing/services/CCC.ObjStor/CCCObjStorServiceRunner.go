package objstor

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

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
	// Get the path to this source file, then navigate to the features directory
	_, filename, _, _ := runtime.Caller(0)
	serviceDir := filepath.Dir(filename)
	featuresPath := filepath.Join(serviceDir, "features")

	runner := &CCCObjStorServiceRunner{}
	runner.AbstractServiceRunner = services.NewAbstractServiceRunner(
		"CCC.ObjStor",
		runner.GetTestResources,
		config,
		featuresPath,
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

	// List all buckets (already includes region information)
	buckets, err := storageService.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
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
