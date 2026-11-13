---
page_title: "platformxe_permissions_override Resource"
description: "Manages a PlatformXe per-user permission override."
---

# platformxe_permissions_override

Manages a per-user permission override that takes precedence over role-based permissions.

~> **One override per key.** PlatformXe keeps exactly one override per (`admin_id`, `path`, `action`); creating one
for a key that already has an override replaces it. Two `platformxe_permissions_override` resources (in one
configuration or in two) for the same key therefore manage the **same** row: the last apply wins, the other drifts on
its next plan, and destroying either deletes the override for both. Declare each key once. The stored `path` is
trimmed and loses any trailing `/`, so `"admin/settings/"` and `"admin/settings"` are the same key.

~> **A live DENY is protected.** Creating a `GRANT` for a key that holds a live (unexpired) `DENY` fails with
`409 OVERRIDE_DENY_ACTIVE`. This resource has no `replace` argument: remove the `DENY` (its resource, or
`DELETE /api/v1/permissions/overrides/remove/{id}`) before applying the `GRANT`. A `DENY` applied over a live `DENY`
keeps the later expiry.

## Example Usage

```hcl
resource "platformxe_permissions_override" "block_admin" {
  admin_id = "usr_123"
  path     = "admin/settings"
  action   = "WRITE"
  effect   = "DENY"
  reason   = "Temporary restriction during audit"
}
```

## Argument Reference

- `admin_id` (String, Required) -- Target user ID.
- `path` (String, Required) -- Permission path.
- `action` (String, Required) -- Action (e.g., `READ`, `WRITE`, `DELETE`).
- `effect` (String, Required) -- `ALLOW` or `DENY`.
- `reason` (String, Optional) -- Human-readable reason for the override.

## Attribute Reference

- `id` (String) -- Override ID, assigned by PlatformXe.

## Import

Permission overrides can be imported using their ID:

```bash
terraform import platformxe_permissions_override.block_admin ovr_abc123
```
