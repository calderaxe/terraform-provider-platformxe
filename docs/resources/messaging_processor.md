---
page_title: "platformxe_messaging_processor Resource"
description: "Manages the PlatformXe messaging processor configuration."
---

# platformxe_messaging_processor

Manages the messaging processor configuration for email, SMS, and WhatsApp dispatch, including provider preferences and channel toggles.

## Example Usage

```hcl
resource "platformxe_messaging_processor" "main" {
  enabled          = true
  email_enabled    = true
  email_from_name  = "MyApp Notifications"
  email_reply_to   = "support@myapp.com"

  sms_enabled        = true
  sms_default_region = "NG"

  whatsapp_enabled = false
}
```

## Argument Reference

- `enabled` (Bool, Optional) -- Whether the messaging processor is enabled. Defaults to `true`.
- `email_enabled` (Bool, Optional) -- Whether email dispatch is enabled. Defaults to `true`.
- `email_preferred_providers` (List of String, Optional) -- Preferred email provider fallback order.
- `email_from_name` (String, Optional) -- Default sender display name for emails.
- `email_reply_to` (String, Optional) -- Default reply-to address for emails.
- `sms_enabled` (Bool, Optional) -- Whether SMS dispatch is enabled. Defaults to `true`.
- `sms_preferred_providers` (List of String, Optional) -- Preferred SMS provider fallback order.
- `sms_default_region` (String, Optional) -- Default region code for SMS delivery.
- `whatsapp_enabled` (Bool, Optional) -- Whether WhatsApp dispatch is enabled. Defaults to `false`.

## Attribute Reference

- `id` (String) -- Processor identifier, assigned by PlatformXe.
- `created_at` (String) -- Timestamp when the processor was created.
- `updated_at` (String) -- Timestamp when the processor was last updated.
