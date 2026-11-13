---
page_title: "platformxe_workflows_trigger Resource"
description: "Manages a PlatformXe event-driven workflow."
---

# platformxe_workflows_trigger

Manages an event-driven workflow with trigger matching and action execution.

## Example Usage

### Event-triggered workflow

```hcl
resource "platformxe_workflows_trigger" "onboarding" {
  name = "New User Onboarding"

  trigger_config {
    trigger_type = "EVENT"
    event_type   = "user.created"
  }

  actions {
    type   = "send_email"
    config = jsonencode({
      template_id = platformxe_template.welcome.id
    })
  }

  actions {
    type   = "webhook"
    config = jsonencode({
      url = "https://example.com/onboard"
    })
  }
}
```

### Scheduled workflow

```hcl
resource "platformxe_workflows_trigger" "daily_report" {
  name        = "Daily Summary Report"
  description = "Sends a daily summary email at 9am."

  trigger_config {
    trigger_type    = "CRON"
    cron_expression = "0 9 * * *"
  }

  actions {
    type   = "send_email"
    config = jsonencode({
      template_id = platformxe_template.daily_summary.id
    })
  }

  priority         = 10
  max_executions_day = 1
}
```

### Conditional workflow

```hcl
resource "platformxe_workflows_trigger" "high_value" {
  name = "High Value Booking Alert"

  trigger_config {
    trigger_type = "EVENT"
    event_type   = "booking.created"
    entity_type  = "BOOKING"
    conditions   = jsonencode({
      all = [
        { attribute = "payload.totalAmount", operator = "gte", value = 100000 }
      ]
    })
  }

  actions {
    type   = "webhook"
    config = jsonencode({
      url = "https://example.com/high-value-alert"
    })
  }

  cooldown_minutes = 5
}
```

## Argument Reference

- `name` (String, Required) -- Workflow name.
- `description` (String, Optional) -- Workflow description.
- `code` (String, Optional) -- Unique workflow code. Auto-generated if not provided.
- `trigger_config` (Block, Required) -- Trigger configuration defining when the workflow fires.
  - `trigger_type` (String, Required) -- Trigger type: EVENT, CRON, or MANUAL.
  - `event_type` (String, Optional) -- Event type to match (required when trigger_type is EVENT).
  - `cron_expression` (String, Optional) -- Cron expression for scheduled triggers (required when trigger_type is CRON).
  - `entity_type` (String, Optional) -- Entity type filter for the trigger.
  - `conditions` (String, Optional) -- JSON-encoded condition object for trigger matching.
- `actions` (Block List, Required) -- Ordered list of actions to execute.
  - `type` (String, Required) -- Action type (e.g., send_email, webhook, delay).
  - `config` (String, Optional) -- JSON-encoded action configuration.
- `is_active` (Boolean, Optional) -- Whether the workflow is active. Defaults to `true`.
- `priority` (Number, Optional) -- Evaluation priority. Higher values take precedence. Defaults to `0`.
- `cooldown_minutes` (Number, Optional) -- Minimum minutes between consecutive executions.
- `max_executions_day` (Number, Optional) -- Maximum executions per day.
- `source_app` (String, Optional) -- Source application identifier.

## Attribute Reference

- `id` (String) -- Workflow ID, assigned by PlatformXe.

## Import

Workflows can be imported using their ID:

```bash
terraform import platformxe_workflows_trigger.onboarding wf_abc123
```
