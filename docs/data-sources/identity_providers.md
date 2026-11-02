---
page_title: "platformxe_identity_providers Data Source"
description: "Lists available identity resolution providers and their health status."
---

# Data Source: platformxe_identity_providers

Lists identity resolution providers and their health status. Use this to check provider availability before configuring identity verification workflows.

## Example Usage

```hcl
data "platformxe_identity_providers" "all" {}

output "provider_names" {
  value = data.platformxe_identity_providers.all.providers[*].name
}

output "healthy_providers" {
  value = [
    for p in data.platformxe_identity_providers.all.providers : p.name
    if p.healthy
  ]
}
```

## Argument Reference

No arguments required.

## Attribute Reference

- `providers` (List) -- List of identity providers with health status.
  - `name` (String) -- Provider name (e.g., smile-id, nimc, dojah).
  - `healthy` (Boolean) -- Whether the provider is currently healthy and accepting requests.
