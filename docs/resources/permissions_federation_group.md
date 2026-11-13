---
page_title: "platformxe_permissions_federation_group Resource"
description: "Manages a PlatformXe federation group for cross-app permissions."
---

# platformxe_permissions_federation_group

Manages a federation group for cross-app permission sharing. Requires an Enterprise plan.

## Example Usage

```hcl
resource "platformxe_permissions_federation_group" "suite" {
  name        = "Caldera Suite"
  description = "Cross-app permission sharing for all Caldera products"
}
```

## Argument Reference

- `name` (String, Required) -- Group name.
- `description` (String, Optional) -- Group description.

## Attribute Reference

- `id` (String) -- Federation group ID, assigned by PlatformXe.

## Import

Federation groups can be imported using their ID:

```bash
terraform import platformxe_permissions_federation_group.suite fg_abc123
```
