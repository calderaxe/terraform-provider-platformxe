---
page_title: "platformxe_qr_processor Resource"
description: "Manages the PlatformXe QR code processor configuration."
---

# platformxe_qr_processor

Manages the QR code processor configuration, including default size, output format, and batch limits.

## Example Usage

```hcl
resource "platformxe_qr_processor" "main" {
  enabled        = true
  default_size   = 400
  default_format = "svg"
  max_batch_size = 50
}
```

## Argument Reference

- `enabled` (Bool, Optional) -- Whether the QR processor is enabled. Defaults to `true`.
- `default_size` (Int64, Optional) -- Default QR code size in pixels. Defaults to `300`.
- `default_format` (String, Optional) -- Default output format. Defaults to `"png"`.
- `max_batch_size` (Int64, Optional) -- Maximum QR codes per batch request. Defaults to `100`.

## Attribute Reference

- `id` (String) -- Processor identifier, assigned by PlatformXe.
- `created_at` (String) -- Timestamp when the processor was created.
- `updated_at` (String) -- Timestamp when the processor was last updated.
