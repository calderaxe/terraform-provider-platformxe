---
page_title: "platformxe_permissions_role Resource"
description: "Manages a PlatformXe permission role."
---

# platformxe_permissions_role

Manages a PlatformXe permission role.

## Example Usage

```hcl
resource "platformxe_permissions_role" "agent" {
  name        = "Support Agent"
  description = "Can view and manage support tickets"
  model       = "SIMPLE"
}
```

## Argument Reference

- `name` (String, Required) -- Role name (1-100 characters).
- `description` (String, Optional) -- Role description.
- `model` (String, Optional) -- Permission model: `SIMPLE` or `FULL`. Defaults to `SIMPLE`. The API cannot change a role's model, so changing it replaces the role.

## Attribute Reference

- `id` (String) -- Role ID, assigned by PlatformXe.

## Import

Roles can be imported using their ID:

```bash
terraform import platformxe_permissions_role.agent role_abc123
```
