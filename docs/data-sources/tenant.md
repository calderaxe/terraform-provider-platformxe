---
page_title: "platformxe_tenant Data Source"
description: "Reads your tenant details based on the API key configured in the provider."
---

# Data Source: platformxe_tenant

Reads your PlatformXe tenant details. The tenant is identified by the API key used in the provider configuration — no ID required.

## Example Usage

```hcl
data "platformxe_tenant" "me" {}

output "tenant_plan" {
  value = data.platformxe_tenant.me.plan
}
```

## Argument Reference

No arguments required. The tenant is resolved from the API key configured in the provider block.

## Attribute Reference

- `id` (String) -- Tenant organization ID.
- `name` (String) -- Tenant display name.
- `slug` (String) -- URL-safe slug.
- `plan` (String) -- Subscription plan: FREE, BASIC, PRO, ENTERPRISE.
- `billing_email` (String) -- Billing contact email.
- `region` (String) -- Region code (e.g., NG).
- `is_active` (Boolean) -- Whether the tenant is active.
