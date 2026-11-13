---
page_title: "platformxe_storage_processor Resource"
description: "Manages the PlatformXe storage processor configuration."
---

# platformxe_storage_processor

Manages the storage processor configuration for media and document file operations, including moderation settings and file size limits.

## Example Usage

```hcl
resource "platformxe_storage_processor" "main" {
  enabled              = true
  provider             = "cloudinary"
  moderation_enabled   = true
  moderation_auto_reject = false

  blocked_mime_types    = ["application/x-executable", "application/x-msdownload"]
  max_image_size_mb     = 15
  max_document_size_mb  = 50
  max_video_size_mb     = 200
}
```

## Argument Reference

- `enabled` (Bool, Optional) -- Whether the storage processor is enabled. Defaults to `true`.
- `provider` (String, Optional) -- Storage provider. Defaults to `"cloudinary"`.
- `moderation_enabled` (Bool, Optional) -- Whether content moderation is enabled. Defaults to `true`.
- `moderation_auto_reject` (Bool, Optional) -- Whether to auto-reject moderated content. Defaults to `false`.
- `blocked_mime_types` (List of String, Optional) -- MIME types to block from upload.
- `max_image_size_mb` (Int64, Optional) -- Maximum image file size in MB. Defaults to `10`.
- `max_document_size_mb` (Int64, Optional) -- Maximum document file size in MB. Defaults to `25`.
- `max_video_size_mb` (Int64, Optional) -- Maximum video file size in MB. Defaults to `100`.

## Attribute Reference

- `id` (String) -- Processor identifier, assigned by PlatformXe.
- `created_at` (String) -- Timestamp when the processor was created.
- `updated_at` (String) -- Timestamp when the processor was last updated.
