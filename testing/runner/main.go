package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
)

var (
	provider       = flag.String("provider", "", "Cloud provider (aws, azure, or gcp)")
	outputDir      = flag.String("output", "", "Output directory for test reports (default: testing/output)")
	timeout        = flag.Duration("timeout", 30*time.Minute, "Timeout for all tests")
	resourceFilter = flag.String("resource", "", "Filter tests to a specific resource name")
	tag            = flag.String("tag", "", "Tag filter to override automatic catalog type filtering (e.g., 'CCC.ObjStor.CN04')")

	// Cloud configuration flags
	region              = flag.String("region", "", "Cloud region")
	azureSubscriptionID = flag.String("azure-subscription-id", "", "Azure subscription ID (required for Azure)")
	azureResourceGroup  = flag.String("azure-resource-group", "", "Azure resource group (required for Azure)")
	azureStorageAccount = flag.String("azure-storage-account", "", "Azure storage account name (required for Azure)")
	gcpProjectID        = flag.String("gcp-project-id", "", "GCP project ID (required for GCP)")
)

func main() {
	flag.Parse()

	// Set default output directory if not specified
	if *outputDir == "" {
		// Get the testing directory (parent of runner directory)
		_, filename, _, _ := runtime.Caller(0)
		runnerDir := filepath.Dir(filename)
		testingDir := filepath.Dir(runnerDir)
		*outputDir = filepath.Join(testingDir, "output")
	}

	// Validate required flags
	if *provider == "" {
		log.Fatal("Error: -provider flag is required (aws, azure, or gcp)")
	}

	if *provider != "aws" && *provider != "azure" && *provider != "gcp" {
		log.Fatalf("Error: invalid provider '%s' (must be aws, azure, or gcp)", *provider)
	}

	// Build CloudParams from flags (priority) or environment variables (fallback)
	cloudParams := buildCloudParams(*provider, *region, *azureSubscriptionID, *azureResourceGroup, *azureStorageAccount, *gcpProjectID)

	// Validate provider-specific requirements
	if err := validateCloudParams(*provider, cloudParams); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Log configuration
	log.Printf("🚀 Starting CCC CFI Compliance Tests")
	log.Printf("   Provider: %s", cloudParams.Provider)
	log.Printf("   Region: %s", cloudParams.Region)
	log.Println()

	// Prepare output directory
	log.Printf("🧹 Cleaning output directory: %s", *outputDir)
	if err := os.RemoveAll(*outputDir); err != nil && !os.IsNotExist(err) {
		log.Printf("⚠️  Warning: Failed to clean output directory: %v", err)
	}
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	log.Printf("✅ Output directory ready")
	log.Println()

	// Assemble list of service runners - one for each service type
	var runners []ServiceRunner
	for _, serviceName := range environment.ServiceTypes {
		runners = append(runners, NewBasicServiceRunner(RunConfig{
			ServiceName:    serviceName,
			CloudParams:    cloudParams,
			OutputDir:      *outputDir,
			Timeout:        *timeout,
			ResourceFilter: *resourceFilter,
			Tag:            *tag,
		}))
	}

	log.Printf("📋 Running %d service runner(s)", len(runners))
	log.Println()

	// Run all service runners
	totalFailed := 0
	totalPassed := 0

	for i, runner := range runners {
		log.Printf("🔧 Running service runner %d/%d", i+1, len(runners))
		exitCode := runner.Run()

		if exitCode == 0 {
			totalPassed++
		} else {
			totalFailed++
		}
	}

	// Combine all OCSF files into a single file
	log.Println("\n🔗 Combining OCSF output files...")
	if err := combineOCSFFiles(*outputDir); err != nil {
		log.Printf("⚠️  Warning: Failed to combine OCSF files: %v", err)
	} else {
		log.Printf("   ✅ Combined OCSF file created: %s", filepath.Join(*outputDir, "combined.ocsf.json"))
	}

	// Print summary
	log.Println("\n" + strings.Repeat("=", 60))
	log.Printf("📊 Overall Summary")
	log.Printf("   Total Runners: %d", len(runners))
	log.Printf("   Passed: %d", totalPassed)
	log.Printf("   Failed: %d", totalFailed)
	log.Println(strings.Repeat("=", 60))

	// Report final results and exit
	if totalFailed > 0 {
		log.Println("❌ Some runners had test failures")
		os.Exit(1)
	} else if len(runners) == 0 {
		log.Println("⚠️  No runners were executed")
		os.Exit(1)
	} else if totalPassed == 0 {
		log.Println("⚠️  No runners executed any tests")
		os.Exit(0) // Not a failure - just nothing to do
	} else {
		log.Println("✅ All runners passed")
		os.Exit(0)
	}
}

// combineOCSFFiles combines all *ocsf.json files in the output directory into a single combined_ocsf.json file
func combineOCSFFiles(outputDir string) error {
	// Find all OCSF JSON files in the output directory
	pattern := filepath.Join(outputDir, "*ocsf.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find OCSF files: %w", err)
	}

	if len(files) == 0 {
		log.Printf("   No OCSF files found to combine")
		return nil
	}

	log.Printf("   Found %d OCSF file(s) to combine", len(files))

	// Combine all JSON arrays into a single array
	var combined []interface{}

	for _, file := range files {
		// Read the file
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("   ⚠️  Warning: Failed to read %s: %v", filepath.Base(file), err)
			continue
		}

		// Parse the JSON array
		var items []interface{}
		if err := json.Unmarshal(data, &items); err != nil {
			log.Printf("   ⚠️  Warning: Failed to parse %s: %v", filepath.Base(file), err)
			continue
		}

		// Add items to the combined array
		combined = append(combined, items...)
		log.Printf("   Added %d item(s) from %s", len(items), filepath.Base(file))
	}

	// Write the combined array to a new file
	combinedPath := filepath.Join(outputDir, "combined.ocsf.json")
	combinedData, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal combined data: %w", err)
	}

	if err := os.WriteFile(combinedPath, combinedData, 0644); err != nil {
		return fmt.Errorf("failed to write combined file: %w", err)
	}

	log.Printf("   Total items in combined file: %d", len(combined))

	return nil
}

// buildCloudParams constructs CloudParams from command-line flags
func buildCloudParams(provider, region, azureSubscriptionID, azureResourceGroup, azureStorageAccount, gcpProjectID string) environment.CloudParams {
	params := environment.CloudParams{
		Provider: provider,
		Region:   region,
	}

	// Only set provider-specific fields
	switch provider {
	case "azure":
		params.AzureSubscriptionID = azureSubscriptionID
		params.AzureResourceGroup = azureResourceGroup
		params.AzureStorageAccount = azureStorageAccount
	case "gcp":
		params.GCPProjectID = gcpProjectID
	case "aws":
		// AWS only needs Provider and Region
	}

	return params
}

// validateCloudParams validates that required parameters are set for the provider
func validateCloudParams(provider string, cloudParams environment.CloudParams) error {
	// Region is required for all providers
	if cloudParams.Region == "" {
		return fmt.Errorf("region is required (use --region flag)")
	}

	// Provider-specific validation
	switch provider {
	case "azure":
		if cloudParams.AzureSubscriptionID == "" {
			return fmt.Errorf("azure subscription ID is required (use --azure-subscription-id flag)")
		}
		if cloudParams.AzureResourceGroup == "" {
			return fmt.Errorf("azure resource group is required (use --azure-resource-group flag)")
		}
	case "gcp":
		if cloudParams.GCPProjectID == "" {
			return fmt.Errorf("GCP project ID is required (use --gcp-project-id flag)")
		}
	}

	return nil
}
