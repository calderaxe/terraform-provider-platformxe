---
page_title: "platformxe_pdf_processor Resource"
description: "Manages the PlatformXe PDF processor configuration."
---

# platformxe_pdf_processor

Manages the PDF processor configuration for document generation, including template selection and branding.

## Example Usage

```hcl
resource "platformxe_pdf_processor" "main" {
  enabled          = true
  default_template = "standard"
  branding_source  = "organization"

  enabled_templates = ["standard", "offer-letter", "property-flyer"]
}
```

## Argument Reference

- `enabled` (Bool, Optional) -- Whether the PDF processor is enabled. Defaults to `true`.
- `default_template` (String, Optional) -- Default PDF template. Defaults to `"standard"`.
- `enabled_templates` (List of String, Optional) -- List of enabled PDF template identifiers.
- `branding_source` (String, Optional) -- Source for branding assets. Defaults to `"organization"`.

## Attribute Reference

- `id` (String) -- Processor identifier, assigned by PlatformXe.
- `created_at` (String) -- Timestamp when the processor was created.
- `updated_at` (String) -- Timestamp when the processor was last updated.
