---
page_title: "platformxe_permissions_federation_member Resource"
description: "Manages a PlatformXe federation group member."
---

# platformxe_permissions_federation_member

Adds an application to a federation group for cross-app permission sharing.

## Example Usage

```hcl
resource "platformxe_permissions_federation_member" "chats" {
  group_id = platformxe_federation_group.suite.id
  app_id   = "app_chats_prod"
}
```

## Argument Reference

- `group_id` (String, Required) -- Federation group ID.
- `app_id` (String, Required) -- Application ID to add to the group.

## Attribute Reference

- `id` (String) -- Member record ID, assigned by PlatformXe.

## Import

Federation members can be imported using their ID:

```bash
terraform import platformxe_permissions_federation_member.chats fm_abc123
```
