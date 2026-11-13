---
page_title: "platformxe_fraud_rule Resource"
description: "Manages a PlatformXe Fraud Detection rule."
---

# platformxe_fraud_rule

Manages a tenant-authored Fraud Detection rule. Rules are predicates evaluated by the Detection Engine on every `decide` call.

Rules move through the lifecycle `draft → shadow → published → archived`. Deleting the resource archives the rule (terminal — fraud rules are append-only on the platform side; "delete" is a soft-archive).

## Example Usage

```hcl
resource "platformxe_fraud_rule" "high_value_ng_withdrawal" {
  name             = "High-value withdrawal in NG"
  status           = "published"
  weight           = 25
  verdict_override = "review"
  description      = "Flag NGN withdrawals over 50k for manual review."

  applies_to_json = jsonencode({
    actions        = ["withdraw"]
    resourceKinds  = ["transaction"]
  })

  condition_json = jsonencode({
    all = [
      { "amount.value" = { gte = 50000 } },
      { "context.geoHint" = { equals = "NG" } },
    ]
  })
}

resource "platformxe_fraud_rule" "velocity_5min" {
  name = "Withdrawal velocity — 5 transactions in 5 minutes"
  status = "shadow"

  applies_to_json = jsonencode({ actions = ["withdraw"] })

  windows_json = jsonencode([
    {
      name        = "withdraws_5m"
      aggregation = "count"
      duration    = "PT5M"
      bucketBy    = "subject.id"
    }
  ])

  condition_json = jsonencode({
    "$count.withdraws_5m" = { gte = 5 }
  })
}
```

## Argument Reference

- `name` (String, Required) — Display name (≤ 200 characters).
- `applies_to_json` (String, Required) — JSON-encoded `appliesTo` object: `{"actions": [...], "resourceKinds": [...]}`.
- `condition_json` (String, Required) — JSON-encoded rule condition AST. Supports the 13-operator DSL (`equals`, `notEquals`, `gt`, `gte`, `lt`, `lte`, `in`, `notIn`, `contains`, `startsWith`, `endsWith`, `exists`, `matches`) plus `all` / `any` / `not` combinators and `$count.<window>` references.
- `status` (String, Optional) — Lifecycle status: `draft`, `shadow`, `published`, `archived`. Defaults to `draft`.
- `weight` (Number, Optional) — Score contribution when the rule triggers. Defaults to `10`.
- `verdict_override` (String, Optional) — Force a specific verdict on trigger regardless of score band: `allow`, `review`, `step_up`, `block`. Used for sanctions hits and similar always-block patterns.
- `description` (String, Optional) — Free-form description.
- `windows_json` (String, Optional) — JSON-encoded array of velocity-counter window definitions. Required when the condition references `$count.<window>`.

## Attribute Reference

- `id` (String) — Rule identifier (`frl_<cuid>`), assigned by PlatformXe.
- `version` (Number) — Server-assigned version, increments on every update.

## Lifecycle

- New rules are created in `draft` status. Setting `status = "published"` (or `"archived"`) at create time triggers a transition immediately after the create call returns.
- Updating `status` triggers the corresponding transition. Note: only `draft → published` and `→ archived` transitions are exposed in v1; the `shadow` state is currently surfaced as read-only by this resource (manage shadow promotion via the API directly when required).
- `terraform destroy` archives the rule. Archive is terminal — to re-introduce, create a new rule.

## Import

```bash
terraform import platformxe_fraud_rule.high_value_ng_withdrawal frl_abc123
```

## Required scope

Manage operations require the `fraud:manage` API key scope.
