---
page_title: "platformxe_threads_channel Resource"
description: "Manages a PlatformXe Contextual Messaging channel."
---

# platformxe_threads_channel

Manages a PlatformXe Contextual Messaging channel -- defines entity type, participant roles, visibility rules, lifecycle automation, and escalation configuration.

## Example Usage

### Basic channel

```hcl
resource "platformxe_threads_channel" "booking" {
  slug               = "booking"
  display_name       = "Booking Conversations"
  entity_type        = "BOOKING"
  participant_roles  = ["GUEST", "HOST", "PLATFORM"]
  default_visibility = ["ALL"]
}
```

### Channel with lifecycle rules

```hcl
resource "platformxe_threads_channel" "booking" {
  slug               = "booking"
  display_name       = "Booking Conversations"
  entity_type        = "BOOKING"
  participant_roles  = ["GUEST", "HOST", "PLATFORM"]
  default_visibility = ["ALL"]

  lifecycle_rules {
    auto_close {
      on_entity_status = ["CHECKED_OUT", "CANCELLED"]
    }

    auto_archive {
      after_closed_days = 90
    }

    system_messages {
      on_thread_created = "A new conversation has been started."
      on_thread_closed  = "This conversation has been closed ({closedReason})."
    }
  }
}
```

### Channel with escalation config

```hcl
resource "platformxe_threads_channel" "booking" {
  slug               = "booking"
  display_name       = "Booking Conversations"
  entity_type        = "BOOKING"
  participant_roles  = ["GUEST", "HOST", "PLATFORM"]
  default_visibility = ["ALL"]

  lifecycle_rules {
    auto_close {
      on_entity_status = ["CHECKED_OUT", "CANCELLED"]
    }
  }

  escalation_config {
    flag_reasons {
      code     = "SAFETY"
      label    = "Safety concern"
      severity = "HIGH"
    }

    flag_reasons {
      code     = "DISPUTE"
      label    = "Payment dispute"
      severity = "MEDIUM"
    }

    flag_reasons {
      code     = "COMPLAINT"
      label    = "General complaint"
      severity = "LOW"
    }

    auto_detection {
      patterns    = ["fire", "flooding", "gas leak", "smoke", "injury"]
      flag_reason = "SAFETY"
      auto_flag   = true
    }

    rules {
      id       = "rule-safety"
      name     = "Safety auto-escalation"
      trigger  = "PARTICIPANT_FLAG"
      conditions = jsonencode({ "in" = [{ "var" = "flag.reason" }, ["SAFETY"]] })
      priority = 1
      is_active = true
      cooldown_minutes = 60

      actions {
        type   = "CREATE_ISSUE"
        config = jsonencode({
          title    = "SAFETY: {{thread.subject}}"
          priority = "URGENT"
          tags     = ["safety", "escalated"]
        })
      }

      actions {
        type   = "NOTIFY_PARTICIPANTS"
        config = jsonencode({
          systemMessage = "A safety concern has been reported."
          roles         = ["ALL"]
        })
      }
    }
  }
}
```

## Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `slug` | String | Yes | Unique channel slug. Forces replacement on change. |
| `display_name` | String | Yes | Human-readable channel name. |
| `entity_type` | String | Yes | Business entity type. Forces replacement on change. |
| `participant_roles` | List(String) | Yes | Allowed participant roles. |
| `default_visibility` | List(String) | Yes | Default message visibility. |
| `lifecycle_rules` | Block | No | Lifecycle automation rules (see below). |
| `escalation_config` | Block | No | Escalation configuration (see below). |
| `webhook_url` | String | No | Channel-level webhook URL. |
| `is_active` | Boolean | No | Whether active. Defaults to `true`. |

### lifecycle_rules

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `auto_close` | Block | No | Automatic thread closure settings. |
| `auto_close.on_entity_status` | List(String) | Yes | Entity statuses that trigger closure. |
| `auto_archive` | Block | No | Automatic thread archival settings. |
| `auto_archive.after_closed_days` | Number | Yes | Days after closure before archival. |
| `system_messages` | Block | No | System message templates. |
| `system_messages.on_thread_created` | String | No | Message on thread creation. |
| `system_messages.on_thread_closed` | String | No | Message on thread closure. Supports `{closedReason}`. |

### escalation_config

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `flag_reasons` | Block List | No | Available flag reasons. |
| `flag_reasons.code` | String | Yes | Reason code (e.g., SAFETY). |
| `flag_reasons.label` | String | Yes | Human-readable label. |
| `flag_reasons.severity` | String | Yes | LOW, MEDIUM, HIGH, or URGENT. |
| `auto_detection` | Block List | No | Auto-detection patterns. |
| `auto_detection.patterns` | List(String) | Yes | Text patterns to detect. |
| `auto_detection.flag_reason` | String | Yes | Flag reason code to apply. |
| `auto_detection.auto_flag` | Boolean | Yes | Whether to auto-flag matches. |
| `rules` | Block List | No | Escalation rules. |
| `rules.id` | String | Yes | Rule identifier. |
| `rules.name` | String | Yes | Rule name. |
| `rules.trigger` | String | Yes | Trigger event (e.g., PARTICIPANT_FLAG). |
| `rules.conditions` | String (JSON) | Yes | JSON Logic condition expression. |
| `rules.actions` | Block List | Yes | Actions to execute. |
| `rules.actions.type` | String | Yes | Action type (CREATE_ISSUE, NOTIFY_PARTICIPANTS, WEBHOOK). |
| `rules.actions.config` | String (JSON) | No | JSON-encoded action configuration. |
| `rules.priority` | Number | Yes | Rule evaluation priority. |
| `rules.is_active` | Boolean | Yes | Whether the rule is active. |
| `rules.cooldown_minutes` | Number | No | Minutes between consecutive rule firings. |

## Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `id` | Channel identifier (assigned by PlatformXe). |

## Import

Thread channels can be imported by ID:

```bash
terraform import platformxe_threads_channel.booking ch_abc123
```

## Notes

- Channels are deactivated on destroy (`is_active = false`), not deleted -- existing threads and messages are preserved.
- `slug` and `entity_type` force replacement because they are part of the channel's identity.
- Escalation rules use [JSON Logic](https://jsonlogic.com) for conditions.
