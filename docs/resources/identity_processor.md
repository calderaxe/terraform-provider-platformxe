---
page_title: "platformxe_identity_processor Resource"
description: "Manages the PlatformXe identity processor configuration."
---

# platformxe_identity_processor

Manages the identity processor configuration for identity resolution and verification, including verification methods and NDPR compliance settings.

## Example Usage

```hcl
resource "platformxe_identity_processor" "main" {
  enabled             = true
  verification_methods = ["bvn", "nin", "passport"]
  consent_required    = true
  ndpr_compliance     = true
}
```

## Argument Reference

- `enabled` (Bool, Optional) -- Whether the identity processor is enabled. Defaults to `true`.
- `verification_methods` (List of String, Optional) -- Allowed identity verification methods.
- `consent_required` (Bool, Optional) -- Whether explicit consent is required before lookups. Defaults to `true`.
- `ndpr_compliance` (Bool, Optional) -- Whether NDPR compliance checks are enforced. Defaults to `true`.

## Attribute Reference

- `id` (String) -- Processor identifier, assigned by PlatformXe.
- `created_at` (String) -- Timestamp when the processor was created.
- `updated_at` (String) -- Timestamp when the processor was last updated.
