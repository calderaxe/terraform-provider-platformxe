---
page_title: "platformxe_webhooks_endpoint Resource"
description: "Manages a PlatformXe outbound webhook endpoint."
---

# platformxe_webhooks_endpoint

Manages an outbound webhook endpoint for event delivery with retry and dead-letter queue.

## Example Usage

```hcl
resource "platformxe_webhooks_endpoint" "notifications" {
  url    = "https://example.com/webhooks"
  events = ["email.sent", "email.bounced", "email.complained"]
  secret = var.webhook_secret
}
```

## Argument Reference

- `url` (String, Required) -- Webhook delivery URL (must be HTTPS).
- `events` (List of String, Required) -- Events to subscribe to.
- `secret` (String, Optional, Sensitive) -- Signing secret for signature verification.

## Attribute Reference

- `id` (String) -- Webhook ID, assigned by PlatformXe.

## Import

Webhooks can be imported using their ID:

```bash
terraform import platformxe_webhooks_endpoint.notifications wh_abc123
```
