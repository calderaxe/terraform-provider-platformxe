---
page_title: "platformxe_events_subscription Resource"
description: "Manages a PlatformXe event subscription."
---

# platformxe_events_subscription

Manages an event subscription: PlatformXe delivers matching events to a webhook URL, signed with the subscription's secret.

## Example Usage

```hcl
resource "platformxe_events_subscription" "email_events" {
  name        = "email-events"
  webhook_url = "https://example.com/events"
  event_types = ["email.sent", "email.delivered"]
}
```

## Argument Reference

- `name` (String, Required) -- Subscription name, at least 3 characters.
- `webhook_url` (String, Required) -- HTTPS URL that receives the events.
- `event_types` (List of String, Required) -- Event types to deliver. At least one.
- `is_active` (Boolean, Optional) -- `true` (default) keeps the subscription ACTIVE; `false` pauses it (status PAUSED). A subscription PlatformXe marked FAILING still reads as active; applying sets it back to ACTIVE.

## Attribute Reference

- `id` (String) -- Subscription ID, assigned by PlatformXe.
- `secret` (String, Sensitive) -- Signing secret. PlatformXe returns it once, at creation, so it is empty after an import. Keep state encrypted.

## Import

Event subscriptions can be imported using their ID:

```bash
terraform import platformxe_events_subscription.email_events sub_abc123
```
