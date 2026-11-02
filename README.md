# Terraform Provider for PlatformXe

Manage [PlatformXe](https://platformxe.com) infrastructure as code — permissions, webhooks, workflows, sending domains, threads channels, fraud rules, identity / messaging / storage / OCR / PDF / QR / exports processors, and the full Custom Events surface (registrations, marketplace listings, federation groups + pushes + external webhook peers).

> **Available on every paid plan.** The provider requires a Basic, Pro, or Enterprise subscription. [Sign up](https://platformxe.com/portal/register) and upgrade to obtain an API key. Some resources are tier-gated (e.g. `platformxe_marketplace_listing` is PRO+, federation resources are ENTERPRISE).

## Install

```hcl
terraform {
  required_providers {
    platformxe = {
      source  = "calderaxe/platformxe"
      version = "~> 1.5.1"
    }
  }
}

provider "platformxe" {
  api_key  = var.platformxe_api_key   # or set the PLATFORMXE_API_KEY env var
  base_url = "https://platformxe.com" # optional — this is the default
}
```

| Attribute  | Type   | Required | Description |
|------------|--------|----------|-------------|
| `api_key`  | string | Yes*     | API key for authentication. Can also be supplied via the `PLATFORMXE_API_KEY` environment variable. |
| `base_url` | string | No       | API base URL. Defaults to `https://platformxe.com`. Override for self-hosted deployments. |

## Resources (26)

### Authorization

| Resource | Description |
|----------|-------------|
| `platformxe_permissions_role` | Permission role (SIMPLE or FULL model) — capabilities or module/action grants |
| `platformxe_permissions_module` | Permission module registration (FULL model) |
| `platformxe_permissions_policy` | ABAC resource policy (path + action + condition + priority + effect) |
| `platformxe_permissions_override` | Per-user permission override (allow/deny) with optional expiry |
| `platformxe_permissions_federation_group` | Cross-app permission federation group (ENTERPRISE) |
| `platformxe_permissions_federation_member` | Member organization within a permission federation group |

### Messaging, content, automations

| Resource | Description |
|----------|-------------|
| `platformxe_webhooks_endpoint` | Outbound webhook endpoint with auto-generated signing secret |
| `platformxe_templates_template` | Content template (subject, HTML/text body, variables) |
| `platformxe_workflows_trigger` | Event-driven workflow automation (trigger config + actions) |
| `platformxe_domains_sending` | Sending domain — emits required DNS records on apply |
| `platformxe_events_subscription` | Event subscription forwarding domain events to a webhook |
| `platformxe_threads_channel` | Caldera Threads channel (slug + lifecycle + escalation rules) |

### Service processors (per-tenant config)

| Resource | Description |
|----------|-------------|
| `platformxe_messaging_processor` | Messaging processor — provider config, fallback rules, branding |
| `platformxe_storage_processor` | Storage processor — Cloudinary + Supabase config, moderation policy |
| `platformxe_ocr_processor` | OCR processor — Azure Computer Vision config |
| `platformxe_pdf_processor` | PDF processor — template registry + asset paths |
| `platformxe_qr_processor` | QR processor — generation defaults |
| `platformxe_identity_processor` | Identity Resolution processor — provider config + circuit breakers |
| `platformxe_exports_processor` | Async exports processor — destination config |

### Fraud Detection

| Resource | Description |
|----------|-------------|
| `platformxe_fraud_rule` | Tenant-authored fraud rule (DSL + lifecycle: draft → shadow → published) |
| `platformxe_fraud_screening_list` | Tenant blocklist / allowlist (entries managed at runtime via SDK) |

### Custom Events (Phase 9A → Pattern 3)

| Resource | Description |
|----------|-------------|
| `platformxe_custom_event` | Tenant-defined custom event — JSON Schema 2020-12 payload, immutable per `(namespace, name, version)` |
| `platformxe_marketplace_listing` | Cross-tenant marketplace listing for a published custom event (PRO+) |
| `platformxe_event_federation_group` | Custom Event Federation group (ENTERPRISE) — distinct from `platformxe_permissions_federation_group` |
| `platformxe_event_federation_push` | Per-version event push declaration on a federation group |
| `platformxe_event_federation_external_peer` | **(v1.5.0 — Pattern 3)** External HTTPS webhook peer, addressed by URL + one-time HMAC secret |

## Data Sources (5)

| Data Source | Description |
|-------------|-------------|
| `platformxe_tenant` | Tenant details resolved from the API key (id, name, slug, plan, region) |
| `platformxe_permissions_modules` | List of registered permission modules |
| `platformxe_identity_providers` | Identity provider health + circuit-breaker status |
| `platformxe_threads_escalation_config` | Resolved escalation config for a threads channel |
| `platformxe_fraud_terms` | Current Fraud Detection terms version + the version this org has accepted |

## Pattern 3 — external webhook peers (1.5.0)

Add an arbitrary HTTPS endpoint as a peer of a Custom Event Federation group. The receiving system doesn't need to be on PlatformXe — relays are POSTed signed with `HMAC-SHA256` against a one-time `whsec_…` secret returned at create time only. Encrypt your state at rest.

```hcl
resource "platformxe_event_federation_group" "partners" {
  name        = "Trusted partners"
  description = "Lettings partners receiving live property events"
}

resource "platformxe_event_federation_external_peer" "bookingcom" {
  group_id    = platformxe_event_federation_group.partners.id
  label       = "Booking.com"
  webhook_url = "https://booking.example.com/inbound/platformxe"

  headers = {
    Authorization = var.bookingcom_inbound_token
  }
}

# `secret` is shown ONLY at create time — store immediately. Rotate by
# destroy + apply.
output "bookingcom_signing_secret" {
  value     = platformxe_event_federation_external_peer.bookingcom.secret
  sensitive = true
}
```

The receiving endpoint verifies inbound POSTs with `X-Caldera-Signature-V2: sha256=<hex>` = HMAC-SHA256(secret, `{X-Caldera-Timestamp}.{raw body}`), a 5-minute window and `X-Caldera-Delivery-Id` as the replay key. The legacy body-only `X-Caldera-Signature` is replayable and is sent only until 2027-04-30, when it switches to the V2 value. See the [federation reference](https://docs.platformxe.com/sdk/federation) for the full wire format and signature-verification examples.

## Worked example

See [`examples/main.tf`](examples/main.tf) for a complete working configuration that exercises the major resources end-to-end.

## Versioning

| Package | Version |
|---------|---------|
| `calderaxe/platformxe` (this) | **1.5.1** |
| `@caldera/platformxe-sdk` (TypeScript) | 1.6.0 *(pending release — SDK MR !5)* |
| `@caldera/platformxe-types` (TypeScript shapes) | 3.2.0 |
| `platformxe` (Python) | 1.6.0 *(pending release — tag `sdk-python/v1.6.0` cut on merge)* |
| `github.com/calderaxe/platformxe-go` (Go) | v1.6.0 *(pending release — tag `sdk-go/v1.6.0` cut on merge)* |

The PlatformXe ecosystem ships **5 published artefacts** that move in lockstep against every API change (the NO-DRIFT policy). The full coverage matrix lives at [docs.platformxe.com/sdk/alignment](https://docs.platformxe.com/sdk/alignment).

## Distribution

The provider is published at `registry.terraform.io/calderaxe/platformxe`. The source of truth lives in the PlatformXe monorepo at `packages/terraform-provider/` and is mirrored to [`github.com/calderaxe/terraform-provider-platformxe`](https://github.com/calderaxe/terraform-provider-platformxe) on every `terraform-provider/v*` tag push. The Terraform Registry resolves new versions from those GitHub tags.

## Development

Build the provider locally:

```bash
cd packages/terraform-provider
go mod tidy
go build -o terraform-provider-platformxe
```

Install the local build for testing:

```bash
PLUGIN_DIR=~/.terraform.d/plugins/registry.terraform.io/calderaxe/platformxe/1.5.1/$(go env GOOS)_$(go env GOARCH)
mkdir -p "$PLUGIN_DIR"
cp terraform-provider-platformxe "$PLUGIN_DIR/"
```

## Documentation

- **Per-resource argument tables + examples:** [docs.platformxe.com/guides/terraform-provider](https://docs.platformxe.com/guides/terraform-provider)
- **Status + incidents:** [status.platformxe.com](https://status.platformxe.com)

## License

Proprietary — © 2026 Caldera Technologies Ltd. All rights reserved.
