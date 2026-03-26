# CCC VPC Feature Authoring Guide Map

Use this file as a repeatable method when creating or revising VPC feature files.

## 1) Purpose

Write scenarios that are:

- clear to read without opening Go code
- traceable to policy intent
- stable in default CI runs
- easy to maintain

## 2) Standard Workflow (Follow in Order)

1. Read the control intent from policy YAML `requirement_text`.
2. Identify scope:
   - subscription/account/region
   - per-resource (VPC/subnet/etc.)
   - runtime behavior
3. Classify test type:
   - `Policy` (configuration/state only)
   - `Behavioural` (runtime observation, non-destructive)
   - `Destructive` (intentional blocked/negative action)
4. Decide run mode:
   - default run (`@CCC.VPC` present)
   - opt-in only (`@OPT_IN`, omit `@CCC.VPC` at scenario level)
5. Define one primary observable for pass/fail.
6. Create scenario with structure: Context -> Action -> Evidence -> Decision.
7. Add optional diagnostic checks only if stable.
8. Validate tag strategy and naming consistency.
9. Run tests and confirm scenario still reads naturally.

## 3) Scenario Structure Pattern

Use this order in each scenario:

- Context:
  - `Given` setup aliases and scope inputs.
- Action:
  - single `When I call ...` for one measurement.
- Evidence:
  - primary observable assertion first.
- Decision:
  - optional `Compliant/ResultClass/Verdict` assertion.

Prefer one main measurement per scenario.

## 4) Assertion Priority

Use assertions in this order:

1. Primary observable (required):
   - boolean (`IsX`)
   - count (`XCount`)
   - membership/status value
2. Decision field (optional):
   - `Compliant`
   - `Verdict`
   - `ResultClass`
3. Reason text (optional):
   - only with stable substring checks

Avoid using adapter-only metadata fields as core correctness checks unless explicitly testing adapter contract.
Examples: `ControlId`, `TestRequirementCategory`.

## 5) Tag Map

Feature-level baseline:

- `@CCC.VPC.CNxx` — control-level tag (enables `--tags '@CCC.VPC.CN01'` filtering)
- `@CCC.VPC.CNxx.ARyy` — assessment requirement tag
- `@tlp-amber` / `@tlp-red`

`@CCC.VPC` placement rule:

- If all scenarios in a file should run by default, place `@CCC.VPC` at feature level.
- If the file mixes default and opt-in scenarios, place `@CCC.VPC` only on default scenarios.

Scenario-level classification:

- `@Policy`
- `@Behavioural`
- `@Destructive`

Scenario-level execution intent:

- `@MAIN` — required on all default CI scenarios (paired with `@DEFAULT`)
- `@DEFAULT`
- `@NEGATIVE`
- `@SANITY`
- `@OPT_IN`

Notes:

- Default VPC service runs rely on `@CCC.VPC` inclusion.
- A scenario without `@CCC.VPC` is excluded from default VPC runs.
- `@MAIN @DEFAULT` appear together on all default scenarios.
- `@Type.P`, `@Type.B`, `@Type.D` are optional and useful only if you filter on them.

## 6) Naming Rules

Use consistent names:

- service handle: `vpcService`
- resource ID aliases: `TargetVpcId`, `SelectedVpcId`
- scenario titles: short, outcome-focused

Avoid ambiguous handle names like `vpc` when it may mean service or resource.

## 7) Supported Step Syntax (Current Library)

Use supported forms:

- equality:
  - `Then "{result}" is "0"`
- boolean:
  - `Then "{result.Compliant}" is true`
- numeric compare:
  - `Then "{result.Count}" should be greater than "0"`
  - `Then "{result.Count}" should be less than "1"`
- contains:
  - `Then "{result.Reason}" contains "expected text"`

Avoid unsupported forms such as:

- `Then "{result}" is greater than 0`
- `Then "{result}" is not empty`

## 8) Templates

### Policy Template

```gherkin
@Policy @MAIN @DEFAULT
@CCC.VPC
Scenario: PASS when <policy condition>
  Given ...
  When I call "{vpcService}" with "<MethodName>" using argument "<param>"
  Then "{result.<PrimaryObservable>}" is "<ExpectedValue>"
```

### Behaviour Template

```gherkin
@Behavioural @DEFAULT
@CCC.VPC
Scenario: PASS when runtime behavior shows <expected behavior>
  Given ...
  When I call "{vpcService}" with "<MethodName>" using argument "<param>"
  Then "{result.<RuntimeObservable>}" is "<ExpectedValue>"
```

### Destructive Template

```gherkin
@Destructive @NEGATIVE @OPT_IN
Scenario: blocked when <disallowed action>
  Given ...
  When I call "{vpcService}" with "<MethodName>" using argument "<param>"
  Then "{result.<BlockObservable>}" is true
```

## 9) Pre-Merge Checklist

Before finalizing a feature file, confirm:

- scenario title states condition + expected outcome
- one primary observable clearly proves pass/fail
- tags reflect both classification and run intent
- no unsupported step syntax
- no unnecessary coupling to non-essential output fields
- default scenarios are stable in normal environments

## 10) Account Prerequisites

Some controls require a one-time manual action in the target AWS account before the default CI scenario will pass. These are not Terraform responsibilities — they reflect account-level state that AWS creates automatically.

### CN01 — Delete the default AWS VPC

AWS automatically creates a default VPC in every region for every account. CN01 checks that no default VPC exists. Until it is manually deleted, **CN01 MAIN will always fail** — this does not indicate broken infrastructure.

```bash
# Check if a default VPC exists
aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --query "Vpcs[*].VpcId"

# Delete subnets inside it first, then the IGW, then the VPC itself
# (AWS will reject the delete if dependent resources remain)
aws ec2 delete-vpc --vpc-id vpc-xxxxxxxx
```

Once deleted, CN01 MAIN passes. CN01 NEGATIVE (which proves detection works) will then fail — that is expected when in the compliant state.

## 11) Simulating Failures (Negative Testing)

Default deployments are compliant by design. To verify detection works, each control should include a `@NEGATIVE @OPT_IN` scenario and a documented way to trigger the failure state.

### CN01 — Default VPC as failure state

CN01 checks for the presence of the AWS account-level default VPC. The compliant state is no default VPC. To simulate failure, leave the default VPC in place (AWS creates one per region automatically).

```bash
# Verify a default VPC exists (failure state)
aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --query "Vpcs[*].VpcId"

# Run negative check
./run-compliance-tests.sh --instance main-aws --service vpc --tags '@NEGATIVE' --output output
```

### CN02 — Redeploy with MapPublicIpOnLaunch=true

CN02 checks that public subnets do not auto-assign external IPs. The default is `false` (compliant). To simulate failure:

```bash
# 1. Set failure state
echo 'map_public_ip_on_launch = true' > remote/aws/vpc/terraform.tfvars
terraform -chdir=remote/aws/vpc apply -auto-approve

# 2. Run negative check
cd testing
./run-compliance-tests.sh --instance main-aws --service vpc --tags '@NEGATIVE' --output output

# 3. Restore compliant state
rm remote/aws/vpc/terraform.tfvars
terraform -chdir=remote/aws/vpc apply -auto-approve
```

### General pattern

Every control should have:

- A `@MAIN @DEFAULT` scenario proving the compliant state passes in normal CI.
- A `@NEGATIVE @OPT_IN` scenario proving the non-compliant state is detected.
- Documentation here (or in the policy README) on how to reach the failure state.

## 12) Definition of Done

A scenario is done when a reviewer can answer all three without reading code:

- What is being tested?
- What evidence proves it?
- Why does that evidence satisfy the control?
