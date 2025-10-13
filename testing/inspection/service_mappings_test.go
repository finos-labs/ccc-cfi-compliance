package inspection

import (
	"testing"
)

func TestGetRegistry(t *testing.T) {
	registry := GetRegistry()
	if registry == nil {
		t.Fatal("GetRegistry returned nil")
	}

	// Verify that we can get the registry multiple times and it's the same instance
	registry2 := GetRegistry()
	if registry != registry2 {
		t.Error("GetRegistry should return the same instance")
	}
}

func TestLookupCatalogType(t *testing.T) {
	tests := []struct {
		provider            string
		providerServiceType string
		expectedCatalogType string
		shouldExist         bool
	}{
		// AWS services
		{"aws", "s3", "CCC.ObjStor", true},
		{"aws", "rds", "CCC.RDMS", true},
		{"aws", "ec2", "CCC.VM", true},
		{"aws", "lambda", "CCC.Serverless", true},
		{"aws", "kms", "CCC.KeyMgmt", true},
		{"aws", "nonexistent", "", false},

		// Azure services
		{"azure", "Microsoft.Storage/storageAccounts", "CCC.ObjStor", true},
		{"azure", "Microsoft.Compute/virtualMachines", "CCC.VM", true},
		{"azure", "Microsoft.DBforPostgreSQL/flexibleServers", "CCC.RDMS", true},
		{"azure", "Microsoft.KeyVault/vaults", "CCC.KeyMgmt", true},
		{"azure", "nonexistent", "", false},

		// GCP services
		{"gcp", "storage.googleapis.com/Bucket", "CCC.ObjStor", true},
		{"gcp", "compute.googleapis.com/Instance", "CCC.VM", true},
		{"gcp", "sqladmin.googleapis.com/Instance", "CCC.RDMS", true},
		{"gcp", "cloudkms.googleapis.com/CryptoKey", "CCC.KeyMgmt", true},
		{"gcp", "nonexistent", "", false},

		// Invalid provider
		{"invalid", "s3", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.provider+"/"+tt.providerServiceType, func(t *testing.T) {
			catalogType, exists := LookupCatalogType(tt.provider, tt.providerServiceType)

			if exists != tt.shouldExist {
				t.Errorf("LookupCatalogType(%q, %q) exists = %v, want %v",
					tt.provider, tt.providerServiceType, exists, tt.shouldExist)
			}

			if exists && catalogType != tt.expectedCatalogType {
				t.Errorf("LookupCatalogType(%q, %q) = %q, want %q",
					tt.provider, tt.providerServiceType, catalogType, tt.expectedCatalogType)
			}
		})
	}
}

func TestGetMapping(t *testing.T) {
	registry := GetRegistry()

	// Test AWS S3 mapping
	mapping, ok := registry.GetMapping("aws", "s3")
	if !ok {
		t.Fatal("Expected to find mapping for aws/s3")
	}

	if mapping.ProviderServiceType != "s3" {
		t.Errorf("ProviderServiceType = %q, want %q", mapping.ProviderServiceType, "s3")
	}

	if mapping.CatalogType != "CCC.ObjStor" {
		t.Errorf("CatalogType = %q, want %q", mapping.CatalogType, "CCC.ObjStor")
	}

	if mapping.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestGetAllMappings(t *testing.T) {
	registry := GetRegistry()

	providers := []string{"aws", "azure", "gcp"}
	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			mappings := registry.GetAllMappings(provider)
			if len(mappings) == 0 {
				t.Errorf("Expected to find mappings for %s, got 0", provider)
			}

			t.Logf("Found %d mappings for %s", len(mappings), provider)
		})
	}

	// Test invalid provider
	mappings := registry.GetAllMappings("invalid")
	if mappings != nil {
		t.Error("Expected nil for invalid provider")
	}
}

func TestAllCatalogTypesUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, catalogType := range AllCatalogTypes {
		if seen[catalogType] {
			t.Errorf("Duplicate catalog type found: %s", catalogType)
		}
		seen[catalogType] = true
	}
}

func TestAllCatalogTypesFormat(t *testing.T) {
	for _, catalogType := range AllCatalogTypes {
		if len(catalogType) < 4 || catalogType[:4] != "CCC." {
			t.Errorf("Catalog type %q should start with 'CCC.'", catalogType)
		}
	}
}
