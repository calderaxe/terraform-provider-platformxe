---
page_title: "platformxe_ocr_processor Resource"
description: "Manages the PlatformXe OCR processor configuration."
---

# platformxe_ocr_processor

Manages the OCR processor configuration for identity document verification, including provider selection, confidence thresholds, and supported document types.

## Example Usage

```hcl
resource "platformxe_ocr_processor" "main" {
  enabled              = true
  provider             = "azure"
  confidence_threshold = 0.90

  supported_document_types = ["passport", "drivers_license", "national_id"]
  languages                = ["en", "fr"]
}
```

## Argument Reference

- `enabled` (Bool, Optional) -- Whether the OCR processor is enabled. Defaults to `true`.
- `provider` (String, Optional) -- OCR provider. Defaults to `"azure"`.
- `confidence_threshold` (Float64, Optional) -- Minimum confidence score to accept OCR results. Defaults to `0.85`.
- `supported_document_types` (List of String, Optional) -- Supported document types for verification.
- `languages` (List of String, Optional) -- Supported languages for OCR processing.

## Attribute Reference

- `id` (String) -- Processor identifier, assigned by PlatformXe.
- `created_at` (String) -- Timestamp when the processor was created.
- `updated_at` (String) -- Timestamp when the processor was last updated.
