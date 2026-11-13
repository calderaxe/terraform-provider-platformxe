---
page_title: "platformxe_fraud_terms Data Source"
description: |-
  Reads the Fraud Detection Engine T&Cs acceptance state for the calling organization.
---

# platformxe_fraud_terms (Data Source)

Reads the calling tenant's Fraud Detection Engine Terms & Conditions acceptance state (Phase 6H).

The fraud cases UI and identity verification endpoints are gated by `requireDetectionPackWithTerms` — a stricter variant of the Detection Pack gate that additionally requires click-through acceptance at the current published version. Use this data source in IaC plans to assert acceptance has been recorded before depending on `fraud:*` or `identity:verify*` workflows.

## Example Usage

```hcl
data "platformxe_fraud_terms" "current" {}

# Fail the plan if T&Cs haven't been accepted at the current version.
check "fraud_terms_accepted" {
  assert {
    condition     = data.platformxe_fraud_terms.current.accepted
    error_message = "Fraud Detection Engine T&Cs are not accepted at the current published version (${data.platformxe_fraud_terms.current.current_version}). Click through at /portal/fraud/terms first."
  }
}
```

## Schema

### Read-Only

- `accepted` (Boolean) — `true` when the org has accepted at the current published version.
- `accepted_version` (String) — Version that was accepted (empty when never accepted).
- `accepted_at` (String) — ISO-8601 timestamp of acceptance.
- `accepted_by` (String) — Tenant user identifier that clicked through.
- `current_version` (String) — The version currently in force.
- `stale_acceptance` (Boolean) — `true` when an acceptance exists but the published version has been bumped — operator must re-accept before strict-gated endpoints clear again.
