---
page_title: "platformxe_permissions_module Resource"
description: "Manages a PlatformXe permission module."
---

# platformxe_permissions_module

Manages a PlatformXe permission module. Modules define groups of permission paths that can be assigned to roles.

**Note:** The API does not support deleting modules. Removing this resource from your configuration will only remove it from Terraform state -- the module remains registered on PlatformXe.

## Example Usage

```hcl
resource "platformxe_permissions_module" "bookings" {
  id          = "LT:BOOKINGS"
  app         = "lettings"
  name        = "Bookings"
  description = "Booking management permissions"
  paths       = ["bookings/*", "reservations/*"]
}
```

## Argument Reference

- `id` (String, Required) -- Module identifier (e.g., LT:BOOKINGS). Forces replacement on change.
- `app` (String, Required) -- Owning application identifier. Forces replacement on change.
- `name` (String, Required) -- Human-readable module name.
- `description` (String, Optional) -- Module description.
- `paths` (List of String, Required) -- Permission paths this module covers.

## Import

Permission modules can be imported using their ID:

```bash
terraform import platformxe_permissions_module.bookings "LT:BOOKINGS"
```
