---
page_title: "platformxe_fraud_screening_list Resource"
description: "Manages a tenant-managed Fraud Detection screening list (blocklist or allowlist)."
---

# platformxe_fraud_screening_list

Manages a tenant-managed screening list. Two kinds are supported: `tenant_blocklist` and `tenant_allowlist`.

The platform's admin-managed lists (`sanctions`, `pep`, `adverse_media`) are NOT manageable from Terraform — they are ingested by the platform's daily refresh cron.

## Example Usage

```hcl
resource "platformxe_fraud_screening_list" "internal_blocklist" {
  source = "internal-2026"
  name   = "Internal blocklist 2026"
  kind   = "tenant_blocklist"
}

resource "platformxe_fraud_screening_list" "vip_allowlist" {
  source = "vip-allowlist"
  name   = "VIP allowlist"
  kind   = "tenant_allowlist"
}
```

## Argument Reference

- `source` (String, Required, ForceNew) — Tenant-defined source identifier, unique within the organisation. Immutable after create — changing this value forces resource replacement.
- `name` (String, Required) — Display name. Mutable in place.
- `kind` (String, Required, ForceNew) — One of `tenant_blocklist` / `tenant_allowlist`. Immutable.

## Attribute Reference

- `id` (String) — List identifier (`slst_<cuid>`), assigned by PlatformXe.

## Managing entries

This resource manages the list itself — the parent record. Entries inside the list are loaded via the SDK's `appendEntries` helper (or the `POST /api/v1/fraud/lists/:id/entries` endpoint). For small static seed lists you can use a `null_resource` with `local-exec` to call `curl` on apply; for larger or dynamic lists, drive the entries from your application code via the SDK.

## Import

```bash
terraform import platformxe_fraud_screening_list.internal_blocklist slst_abc123
```

## Required scope

Manage operations require the `fraud:manage` API key scope.
