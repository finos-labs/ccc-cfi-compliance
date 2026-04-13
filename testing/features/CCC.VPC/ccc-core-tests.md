# CCC.Core Controls — VPC Test Coverage

CCC.VPC inherits the following Core controls: CN01, CN03, CN04, CN05, CN06, CN07, CN09.
Service tag used to include a Core scenario in VPC runs: `@vpc`

---

## Coverage Status

| Control | Description | Need | Reason | Status |
| --- | --- | --- | --- | --- |
| CN01.AR01 | TLS 1.3 enforcement | Yes | VPC endpoints must enforce TLS — @vpc already on all AR01 scenarios | Applied |
| CN01.AR02–AR05 | Further TLS / encryption controls | Yes | Explicitly inherited — needs @vpc scenarios per AR | Pending |
| CN03.AR01–AR04 | MFA for destructive / admin operations | Yes | Explicitly inherited — VPC deletion/modification requires MFA policy check | Pending |
| CN04.AR01 | Admin access logging | Yes | VPC API calls logged via CloudTrail | Applied |
| CN04.AR02 | Data modification logging | Yes | VPC config changes (route tables, SGs) logged via CloudTrail | Applied |
| CN04.AR03 | Data read logging | Yes | VPC describe/read calls captured in CloudTrail | Applied |
| CN05.AR01 | Block unrestricted public ingress | Yes | VPC-level public access restriction | Applied |
| CN05.AR02–AR06 | Further access controls | Yes | Explicitly inherited — needs @vpc scenarios; behavioural ARs need VPC-specific steps | Pending |
| CN06.AR01 | Resource region compliance | Yes | VPC must be in approved region | Applied |
| CN06.AR02 | Child resource region compliance | No | Subnets inherit VPC region — NotTestable for VPC | — |
| CN07.AR01 | Enumeration event publishing | Yes | VPC-level network enumeration detection | Applied |
| CN07.AR02 | Enumeration activity logging | Yes | VPC API enumeration logged via CloudTrail | Applied |
| CN09.AR01 | Log separation | Yes | Flow logs must be stored separately from the VPC | Applied |
| CN09.AR02 | Logs cannot be disabled | No | Requires behavioural runtime test — out of scope for now | — |
| CN09.AR03 | Log redirection requires halt | No | Requires behavioural runtime test — out of scope for now | — |

---

## Notes

- All applied scenarios are `@Policy` checks — no behavioural side effects.
- CN04.AR02 uses `data-write-logging` policy check (CloudTrail write events cover VPC config changes).
- CN07 policy checks (`enumeration-monitoring-policy`, `enumeration-logging-policy`) apply to VPC API enumeration via CloudTrail.
- CN09.AR01 existing `@vpc` scenario renamed to reflect VPC context (flow log separation).
- CN05.AR02–AR06 deferred: some scenarios have behavioural components that need VPC-specific step implementations before tagging.
