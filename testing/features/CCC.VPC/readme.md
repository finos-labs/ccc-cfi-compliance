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
   - `Behavior` (runtime observation, non-destructive)
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

- `@CCC.VPC.CNxx.ARyy`
- `@tlp-amber` / `@tlp-red`

`@CCC.VPC` placement rule:

- If all scenarios in a file should run by default, place `@CCC.VPC` at feature level.
- If the file mixes default and opt-in scenarios, place `@CCC.VPC` only on default scenarios.

Scenario-level classification:

- `@Policy`
- `@Behavior`
- `@Destructive`

Scenario-level execution intent:

- `@DEFAULT`
- `@NEGATIVE`
- `@SANITY`
- `@OPT_IN`

Notes:

- Default VPC service runs rely on `@CCC.VPC` inclusion.
- A scenario without `@CCC.VPC` is excluded from default VPC runs.
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
@Policy @DEFAULT
@CCC.VPC
Scenario: PASS when <policy condition>
  Given ...
  When I call "{vpcService}" with "<MethodName>" <with parameter ...>
  Then "{result.<PrimaryObservable>}" is "<ExpectedValue>"
```

### Behavior Template

```gherkin
@Behavior @DEFAULT
@CCC.VPC
Scenario: PASS when runtime behavior shows <expected behavior>
  Given ...
  When I call "{vpcService}" with "<MethodName>" <with parameter ...>
  Then "{result.<RuntimeObservable>}" is "<ExpectedValue>"
```

### Destructive Template

```gherkin
@Destructive @NEGATIVE @OPT_IN
Scenario: blocked when <disallowed action>
  Given ...
  When I call "{vpcService}" with "<MethodName>" <with parameter ...>
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

## 10) Definition of Done

A scenario is done when a reviewer can answer all three without reading code:

- What is being tested?
- What evidence proves it?
- Why does that evidence satisfy the control?
