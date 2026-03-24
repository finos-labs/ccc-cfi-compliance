package vpc

// CN02Service covers CCC.VPC.CN02: no auto external IP assignment on public subnets.
type CN02Service interface {
	ListPublicSubnets(vpcID string) ([]interface{}, error)
	SummarizePublicSubnets(vpcID string) (string, error)
	EvaluatePublicSubnetDefaultIPControl(vpcID string) (map[string]interface{}, error)
	SelectPublicSubnetForTest(vpcID string) (map[string]interface{}, error)
	CreateTestResourceInSubnet(subnetID string) (map[string]interface{}, error)
	GetResourceExternalIpAssignment(resourceID string) (map[string]interface{}, error)
	DeleteTestResource(resourceID string) (map[string]interface{}, error)
}
