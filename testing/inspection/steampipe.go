package inspection

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// GetAccessiblePorts queries Steampipe for accessible ports based on cloud provider
func GetAccessiblePorts(ctx context.Context, provider string) ([]TestParams, error) {
	// Validate provider first before connecting
	var query string
	switch provider {
	case "aws":
		query = getAWSPortsQuery()
	case "azure":
		query = getAzurePortsQuery()
	case "gcp":
		query = getGCPPortsQuery()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	db, err := connectSteampipe()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Steampipe: %w", err)
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query ports: %w", err)
	}
	defer rows.Close()

	return parsePortResults(rows, provider)
}

// GetServices queries Steampipe for cloud services based on provider
func GetServices(ctx context.Context, provider string) ([]TestParams, error) {
	// Validate provider first before connecting
	var query string
	switch provider {
	case "aws":
		query = getAWSServicesQuery()
	case "azure":
		query = getAzureServicesQuery()
	case "gcp":
		query = getGCPServicesQuery()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	db, err := connectSteampipe()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Steampipe: %w", err)
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	return parseServiceResults(rows, provider)
}

// connectSteampipe establishes a connection to the Steampipe PostgreSQL database
func connectSteampipe() (*sql.DB, error) {
	// Steampipe runs on localhost:9193 by default
	connStr := "host=localhost port=9193 user=steampipe dbname=steampipe sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Steampipe: %w", err)
	}

	return db, nil
}

// getAWSPortsQuery returns SQL to find accessible ports in AWS
func getAWSPortsQuery() string {
	return `
		SELECT DISTINCT
			sg.group_id as uid,
			sg.group_name as service_type,
			ip_permission->>'FromPort' as port_number,
			ip_permission->>'IpProtocol' as protocol,
			sg.region,
			sg.arn as resource_arn,
			'open' as status,
			sg.tags::text as labels,
			i.public_ip_address as hostname
		FROM 
			aws_vpc_security_group sg,
			jsonb_array_elements(ip_permissions) as ip_permission
		INNER JOIN aws_ec2_instance i ON sg.group_id = ANY(
			SELECT jsonb_array_elements_text(
				jsonb_path_query_array(i.security_groups, '$[*].GroupId')
			)
		)
		WHERE 
			(ip_permission->'IpRanges' @> '[{"CidrIp": "0.0.0.0/0"}]'
			 OR ip_permission->'Ipv6Ranges' @> '[{"CidrIpv6": "::/0"}]')
			AND i.public_ip_address IS NOT NULL
			AND i.public_ip_address != ''
		ORDER BY 
			sg.region, port_number;
	`
}

// getAzurePortsQuery returns SQL to find accessible ports in Azure
func getAzurePortsQuery() string {
	return `
		SELECT DISTINCT
			nsg.id as uid,
			nsg.name as service_type,
			rule->>'destinationPortRange' as port_number,
			rule->>'protocol' as protocol,
			nsg.region,
			nsg.id as resource_arn,
			'open' as status,
			nsg.tags::text as labels,
			vm.public_ip_address as hostname
		FROM 
			azure_network_security_group nsg,
			jsonb_array_elements(security_rules) as rule,
			azure_compute_virtual_machine vm
		WHERE 
			rule->>'access' = 'Allow'
			AND rule->>'direction' = 'Inbound'
			AND (rule->>'sourceAddressPrefix' = '*' OR rule->>'sourceAddressPrefix' = 'Internet')
			AND vm.network_profile->>'networkInterfaces' LIKE '%' || nsg.id || '%'
			AND vm.public_ip_address IS NOT NULL
			AND vm.public_ip_address != ''
		ORDER BY 
			nsg.region, port_number;
	`
}

// getGCPPortsQuery returns SQL to find accessible ports in GCP
func getGCPPortsQuery() string {
	return `
		SELECT DISTINCT
			fw.id::text as uid,
			fw.name as service_type,
			COALESCE(allowed_rule->>'ports', 'all') as port_number,
			COALESCE(allowed_rule->>'IPProtocol', 'all') as protocol,
			COALESCE(fw.location, 'global') as region,
			fw.self_link as resource_arn,
			'open' as status,
			'{}'::text as labels,
			COALESCE(
				ac->>'natIP',
				ac->>'externalIpv6'
			) as hostname
		FROM 
			gcp_compute_firewall fw,
			jsonb_array_elements(allowed) as allowed_rule,
			jsonb_array_elements_text(source_ranges) as source_range,
			gcp_compute_instance i,
			jsonb_array_elements(i.network_interfaces) as ni,
			jsonb_array_elements(ni->'accessConfigs') as ac
		WHERE 
			source_range = '0.0.0.0/0'
			AND direction = 'INGRESS'
			AND ni->>'network' = fw.network
			AND (ac->>'natIP' IS NOT NULL OR ac->>'externalIpv6' IS NOT NULL)
		ORDER BY 
			region, port_number;
	`
}

// getAWSServicesQuery returns SQL to find AWS services using the tagging resource table
// This provides universal discovery of all tagged resources across service types
func getAWSServicesQuery() string {
	return `
		SELECT 
			split_part(arn, ':', 3) as service_type,
			name as service_name,
			arn as endpoint,
			region,
			arn as resource_arn,
			'active' as status,
			tags::text as labels,
			arn as uid
		FROM aws_tagging_resource
		WHERE region != ''
		AND region IS NOT NULL
		ORDER BY service_type, name;
	`
}

// getAzureServicesQuery returns SQL to find Azure services using the resource table
// This provides universal discovery of all resources across service types
func getAzureServicesQuery() string {
	return `
		SELECT 
			type as service_type,
			name as service_name,
			id as endpoint,
			region,
			id as resource_arn,
			COALESCE(provisioning_state, 'active') as status,
			tags::text as labels,
			id as uid
		FROM azure_resource
		WHERE region IS NOT NULL
		AND region != ''
		ORDER BY type, name;
	`
}

// getGCPServicesQuery returns SQL to find GCP services using the cloud asset table
// This provides universal discovery of all resources across service types
// Requires Cloud Asset API to be enabled: https://console.cloud.google.com/apis/library/cloudasset.googleapis.com
func getGCPServicesQuery() string {
	return `
		SELECT 
			asset_type as service_type,
			split_part(name, '/', array_length(string_to_array(name, '/'), 1)) as service_name,
			name as endpoint,
			COALESCE(resource->>'location', project) as region,
			name as resource_arn,
			'active' as status,
			COALESCE((resource->'labels')::text, '{}') as labels,
			name as uid
		FROM gcp_cloud_asset
		WHERE asset_type NOT LIKE '%/projects/%'
		AND asset_type NOT LIKE '%/organizations/%'
		ORDER BY asset_type, name;
	`
}

// parsePortResults converts SQL rows to TestParams structs
func parsePortResults(rows *sql.Rows, provider string) ([]TestParams, error) {
	var ports []TestParams

	for rows.Next() {
		var port TestParams
		var labelsJSON sql.NullString
		var resourceARN string // Captured but not stored in TestParams
		var status string      // Captured but not stored in TestParams
		var hostname string    // Hostname/IP from query

		err := rows.Scan(
			&port.UID,
			&port.ServiceType,
			&port.PortNumber,
			&port.Protocol,
			&port.Region,
			&resourceARN,
			&status,
			&labelsJSON,
			&hostname,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		port.Provider = provider
		port.HostName = hostname
		port.ResourceName = port.ServiceType // For ports, use the security group/firewall name

		// Set ProviderServiceType from ServiceType for now (will be properly extracted in future)
		port.ProviderServiceType = port.ServiceType

		// Lookup CatalogType from service mapping
		if catalogType, ok := LookupCatalogType(provider, port.ProviderServiceType); ok {
			port.CatalogType = catalogType
		} else {
			// If no mapping found, leave CatalogType empty
			port.CatalogType = ""
		}

		// Parse labels from JSON string if present
		if labelsJSON.Valid {
			// Simple parsing - in production you'd want proper JSON unmarshaling
			port.Labels = []string{labelsJSON.String}
		}

		ports = append(ports, port)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ports, nil
}

// parseServiceResults converts SQL rows to TestParams structs
func parseServiceResults(rows *sql.Rows, provider string) ([]TestParams, error) {
	var services []TestParams

	for rows.Next() {
		var svc TestParams
		var labelsJSON sql.NullString
		var serviceName string // Resource name for display
		var endpoint string    // Captured but not stored in TestParams
		var resourceARN string // Captured but not stored in TestParams
		var status string      // Captured but not stored in TestParams

		err := rows.Scan(
			&svc.ServiceType,
			&serviceName,
			&endpoint, // Endpoint/resource identifier (not stored in TestParams)
			&svc.Region,
			&resourceARN,
			&status,
			&labelsJSON,
			&svc.UID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		svc.Provider = provider
		svc.HostName = ""              // Services don't have hostnames
		svc.PortNumber = ""            // Services don't have specific ports
		svc.Protocol = ""              // Services don't have specific protocols
		svc.ResourceName = serviceName // Human-readable resource name

		// Set ProviderServiceType from ServiceType for now (will be properly extracted in future)
		svc.ProviderServiceType = svc.ServiceType

		// Lookup CatalogType from service mapping
		if catalogType, ok := LookupCatalogType(provider, svc.ProviderServiceType); ok {
			svc.CatalogType = catalogType
		} else {
			// If no mapping found, leave CatalogType empty
			svc.CatalogType = ""
		}

		// Parse labels from JSON string if present
		if labelsJSON.Valid {
			svc.Labels = []string{labelsJSON.String}
		}

		services = append(services, svc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return services, nil
}
