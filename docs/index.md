---
page_title: "PlatformXe Provider"
description: "The PlatformXe provider manages resources on the PlatformXe infrastructure platform."
---

# PlatformXe Provider

> **Paid tenants only.** The Terraform provider is available to Basic, Pro, and Enterprise plan subscribers. [Sign up](https://platformxe.com/portal/register) and upgrade your plan to get access.

The PlatformXe provider allows you to manage [PlatformXe](https://platformxe.com) resources as infrastructure-as-code. All resource names follow the `platformxe_{service}_{resource}` pattern.

## Resources

| Resource | Service | Description |
|----------|---------|-------------|
| `platformxe_permissions_role` | Permissions | Permission roles (SIMPLE or FULL model) |
| `platformxe_permissions_policy` | Permissions | ABAC resource policies with condition logic |
| `platformxe_permissions_override` | Permissions | Per-user permission overrides |
| `platformxe_permissions_module` | Permissions | Permission modules (path groups for roles) |
| `platformxe_permissions_federation_group` | Permissions | Federation groups for cross-app permissions |
| `platformxe_permissions_federation_member` | Permissions | Federation group members |
| `platformxe_webhooks_endpoint` | Webhooks | Outbound webhook endpoints |
| `platformxe_templates_template` | Templates | Email/message templates |
| `platformxe_workflows_trigger` | Workflows | Event-driven workflow automations |
| `platformxe_domains_sending` | Domains | Sending domains for email delivery |
| `platformxe_events_subscription` | Events | Event subscriptions |
| `platformxe_threads_channel` | Threads | Contextual Messaging channels |
| `platformxe_ocr_processor` | OCR | OCR processor configuration |
| `platformxe_pdf_processor` | PDF | PDF processor configuration |
| `platformxe_qr_processor` | QR | QR code processor configuration |
| `platformxe_messaging_processor` | Messaging | Messaging processor configuration |
| `platformxe_storage_processor` | Storage | Storage processor configuration |
| `platformxe_exports_processor` | Exports | Exports processor configuration |
| `platformxe_identity_processor` | Identity | Identity processor configuration |

## Data Sources

| Data Source | Service | Description |
|-------------|---------|-------------|
| `platformxe_tenant` | Global | Tenant details (resolved from API key) |
| `platformxe_permissions_modules` | Permissions | List all registered permission modules |
| `platformxe_identity_providers` | Identity | Identity provider health status |
| `platformxe_threads_escalation_config` | Threads | Channel escalation configuration |

## Example Usage

```hcl
terraform {
  required_providers {
    platformxe = {
      source  = "calderaxe/platformxe"
      version = "~> 1.0"
    }
  }
}

provider "platformxe" {
  api_key = var.platformxe_api_key
}

resource "platformxe_permissions_role" "admin" {
  name  = "Admin"
  model = "SIMPLE"
}

resource "platformxe_threads_channel" "booking" {
  slug               = "booking"
  display_name       = "Booking Conversations"
  entity_type        = "BOOKING"
  participant_roles  = ["GUEST", "HOST", "PLATFORM"]
  default_visibility = ["ALL"]
}
```

## Authentication

The provider requires an API key for authentication. You can provide it in the provider configuration or via the `PLATFORMXE_API_KEY` environment variable.

```bash
export PLATFORMXE_API_KEY="pxk_live_your_key_here"
```

## Schema

### Optional

- `api_key` (String, Sensitive) -- PlatformXe API key. Can also be set via `PLATFORMXE_API_KEY` environment variable.
- `base_url` (String) -- API base URL. Defaults to `https://platformxe.com`.
