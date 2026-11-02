---
page_title: "platformxe_threads_escalation_config Data Source"
description: "Reads the escalation configuration for a PlatformXe Contextual Messaging channel."
---

# Data Source: platformxe_threads_escalation_config

Reads the escalation configuration for a PlatformXe Contextual Messaging channel. Use this to inspect flag reasons, auto-detection patterns, and escalation rules.

## Example Usage

```hcl
data "platformxe_threads_escalation_config" "booking" {
  channel_id = platformxe_threads_channel.booking.id
}

output "flag_reason_codes" {
  value = data.platformxe_threads_escalation_config.booking.flag_reasons[*].code
}
```

## Argument Reference

- `channel_id` (String, Required) -- The channel ID to read escalation config from.

## Attribute Reference

- `flag_reasons` (List) -- Available flag reasons for the channel.
  - `code` (String) -- Flag reason code (e.g., SAFETY, DISPUTE).
  - `label` (String) -- Human-readable label.
  - `severity` (String) -- Severity level: LOW, MEDIUM, HIGH, or URGENT.
- `auto_detection` (List) -- Auto-detection patterns configured for the channel.
  - `patterns` (List of String) -- Text patterns to detect in messages.
  - `flag_reason` (String) -- Flag reason code applied when patterns match.
  - `auto_flag` (Boolean) -- Whether matching messages are automatically flagged.
- `rules` (List) -- Escalation rules configured for the channel.
  - `id` (String) -- Rule identifier.
  - `name` (String) -- Rule name.
  - `trigger` (String) -- Trigger event (e.g., PARTICIPANT_FLAG).
  - `conditions` (String) -- JSON Logic condition expression.
  - `actions` (List) -- Actions executed when the rule fires.
    - `type` (String) -- Action type (CREATE_ISSUE, NOTIFY_PARTICIPANTS, WEBHOOK).
    - `config` (String) -- JSON-encoded action configuration.
  - `priority` (Number) -- Rule evaluation priority.
  - `is_active` (Boolean) -- Whether the rule is active.
  - `cooldown_minutes` (Number) -- Minutes between consecutive rule firings.
