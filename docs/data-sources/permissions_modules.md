---
page_title: "platformxe_permissions_modules Data Source"
description: "Lists all registered permission modules."
---

# Data Source: platformxe_permissions_modules

Lists all registered permission modules across the PlatformXe ecosystem.

## Example Usage

```hcl
data "platformxe_permissions_modules" "all" {}

output "module_count" {
  value = length(data.platformxe_permissions_modules.all.modules)
}
```

## Attribute Reference

- `modules` (List) -- List of permission modules.
  - `id` (String) -- Module ID.
  - `app` (String) -- Source application.
  - `name` (String) -- Module name.
  - `description` (String) -- Module description.
