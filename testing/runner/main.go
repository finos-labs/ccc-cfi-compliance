package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/finos-labs/ccc-cfi-compliance/testing/services"
	objstor "github.com/finos-labs/ccc-cfi-compliance/testing/services/CCC.ObjStor"
)

var (
	provider       = flag.String("provider", "", "Cloud provider (aws, azure, or gcp)")
	outputDir      = flag.String("output", "output", "Output directory for test reports")
	timeout        = flag.Duration("timeout", 30*time.Minute, "Timeout for all tests")
	resourceFilter = flag.String("resource", "", "Filter tests to a specific resource name")
)

func main() {
	flag.Parse()

	// Validate required flags
	if *provider == "" {
		log.Fatal("Error: -provider flag is required (aws, azure, or gcp)")
	}

	if *provider != "aws" && *provider != "azure" && *provider != "gcp" {
		log.Fatalf("Error: invalid provider '%s' (must be aws, azure, or gcp)", *provider)
	}

	log.Printf("🚀 Starting CCC CFI Compliance Tests")
	log.Printf("   Provider: %s", *provider)
	log.Printf("   Output Directory: %s", *outputDir)
	log.Printf("   Timeout: %s", *timeout)
	if *resourceFilter != "" {
		log.Printf("   Resource Filter: %s", *resourceFilter)
	}
	log.Println()

	// Build shared configuration
	baseConfig := services.RunConfig{
		Provider:       *provider,
		OutputDir:      *outputDir,
		Timeout:        *timeout,
		ResourceFilter: *resourceFilter,
	}

	// Assemble list of service runners
	runners := []services.ServiceRunner{
		objstor.NewCCCObjStorServiceRunner(baseConfig),
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
		log.Println("❌ Some runners failed")
		os.Exit(1)
	} else if len(runners) == 0 {
		log.Println("⚠️  No runners were executed")
		os.Exit(1)
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
