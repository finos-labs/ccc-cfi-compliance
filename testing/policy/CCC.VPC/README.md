# CCC.VPC - Network Security Controls

This folder contains policy reference material for CCC `VPC` (virtual network) controls.

## Structure

- Control documentation: `CCC.VPC.CNxx/README.md`
- Provider-specific evidence/query definitions: `CCC.VPC.CNxx/<provider>/ARyy/*.yaml`
  - These YAML files document *how* a requirement can be evaluated using control-plane evidence (e.g., cloud CLI/API responses).
  - The executable compliance checks live under `testing/features/` and `testing/api/`.

## Mapping to executable tests

For AWS, the VPC service implementation is in `testing/api/vpc/` and is exposed to tests via `GetServiceAPI("vpc")` in `testing/api/factory/aws_factory.go`.

Feature tests for these controls live under `testing/features/CCC.VPC/` and are selected by tags like `@CCC.VPC.CN01.AR01`.

