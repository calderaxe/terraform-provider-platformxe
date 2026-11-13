---
page_title: "platformxe_exports_processor Resource"
description: "Manages the PlatformXe exports processor configuration."
---

# platformxe_exports_processor

Manages the exports processor configuration for data export jobs, including allowed formats, row limits, and retention policies.

## Example Usage

```hcl
resource "platformxe_exports_processor" "main" {
  enabled            = true
  allowed_formats    = ["csv", "json", "xlsx"]
  max_rows_per_export = 50000
  retention_days     = 14
}
```

## Argument Reference

- `enabled` (Bool, Optional) -- Whether the exports processor is enabled. Defaults to `true`.
- `allowed_formats` (List of String, Optional) -- Allowed export formats. Defaults to `["csv", "json"]`.
- `max_rows_per_export` (Int64, Optional) -- Maximum rows per export job. Defaults to `100000`.
- `retention_days` (Int64, Optional) -- Days to retain export files. Defaults to `30`.

## Attribute Reference

- `id` (String) -- Processor identifier, assigned by PlatformXe.
- `created_at` (String) -- Timestamp when the processor was created.
- `updated_at` (String) -- Timestamp when the processor was last updated.
