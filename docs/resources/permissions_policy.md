---
page_title: "platformxe_permissions_policy Resource"
description: "Manages a PlatformXe ABAC permission policy."
---

# platformxe_permissions_policy

Manages a PlatformXe ABAC permission policy with condition-based access control.

## Example Usage

### Single condition

```hcl
resource "platformxe_permissions_policy" "eu_data" {
  path   = "data/*"
  action = "read"
  effect = "ALLOW"

  condition {
    combinator = "all"

    rules {
      attribute = "user.region"
      operator  = "eq"
      value     = "EU"
    }
  }
}
```

### Multiple conditions

```hcl
resource "platformxe_permissions_policy" "admin_only" {
  path        = "settings/*"
  action      = "write"
  effect      = "ALLOW"
  description = "Only admins can write settings during business hours."
  priority    = 10

  condition {
    combinator = "all"

    rules {
      attribute = "user.role"
      operator  = "eq"
      value     = "admin"
    }

    rules {
      attribute = "context.hour"
      operator  = "gte"
      value     = "9"
    }

    rules {
      attribute = "context.hour"
      operator  = "lte"
      value     = "17"
    }
  }
}
```

### Deny policy with array operator

```hcl
resource "platformxe_permissions_policy" "block_regions" {
  path   = "exports/*"
  action = "read"
  effect = "DENY"

  condition {
    combinator = "any"

    rules {
      attribute = "user.region"
      operator  = "in"
      value     = jsonencode(["SANCTIONED_A", "SANCTIONED_B"])
    }
  }
}
```

## Argument Reference

- `path` (String, Required) -- Resource path pattern the policy applies to. Supports wildcards (`*`).
- `action` (String, Required) -- Action the policy governs (e.g., read, write, delete).
- `effect` (String, Required) -- `ALLOW` or `DENY`.
- `condition` (Block, Optional) -- ABAC condition expression with combinator logic.
  - `combinator` (String, Required) -- Logic combinator: `all`, `any`, or `not`.
  - `rules` (Block List, Required) -- Condition rules to evaluate.
    - `attribute` (String, Required) -- Attribute path (e.g., `user.region`, `resource.owner`).
    - `operator` (String, Required) -- Comparison operator: `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `in`, `not_in`, `contains`, `starts_with`, `ends_with`, `exists`, `not_exists`.
    - `value` (String, Required) -- Value to compare against. For array operators (`in`, `not_in`), use `jsonencode()`.
- `priority` (Number, Optional) -- Evaluation priority. Higher values take precedence. Defaults to `0`.
- `description` (String, Optional) -- Policy description.
- `is_active` (Boolean, Optional) -- Whether the policy is active. Defaults to `true`.

## Attribute Reference

- `id` (String) -- Policy ID, assigned by PlatformXe.

## Import

Resource policies can be imported using their ID:

```bash
terraform import platformxe_permissions_policy.eu_data pol_abc123
```
