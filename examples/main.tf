terraform {
  required_providers {
    platformxe = {
      source = "calderasuite/platformxe"
    }
  }
}

provider "platformxe" {
  api_key = var.platformxe_api_key
  # base_url = "https://platformxe.com"  # default
}

variable "platformxe_api_key" {
  type      = string
  sensitive = true
}

# --- Tenant ---

# Your tenant details (derived from API key)
data "platformxe_tenant" "me" {}

output "tenant_plan" {
  value = data.platformxe_tenant.me.plan
}

# --- Data Sources ---

# List all registered permission modules
data "platformxe_permissions_modules" "all" {}

# Check identity provider health
data "platformxe_identity_providers" "health" {}

# --- Roles ---

resource "platformxe_permissions_role" "agent" {
  name        = "Support Agent"
  description = "Can view and manage support tickets"
  model       = "SIMPLE"
}

resource "platformxe_permissions_role" "manager" {
  name        = "Team Manager"
  description = "Full access to team resources"
  model       = "FULL"
}

# --- Permission Module ---

resource "platformxe_permissions_module" "bookings" {
  id          = "LT:BOOKINGS"
  app         = "lettings"
  name        = "Bookings"
  description = "Booking management permissions"
  paths       = ["bookings/*", "reservations/*"]
}

# --- Resource Policy ---

resource "platformxe_permissions_policy" "deny_delete_invoices" {
  path        = "invoices/*"
  action      = "delete"
  effect      = "DENY"
  priority    = 100
  description = "Prevent deletion of invoices"
}

# --- Permission Override ---

resource "platformxe_permissions_override" "temp_access" {
  admin_id   = "admin_abc123"
  path       = "reports/*"
  action     = "read"
  effect     = "ALLOW"
  reason     = "Temporary audit access"
  expires_at = "2026-06-01T00:00:00Z"
}

# --- Webhook ---

resource "platformxe_webhooks_endpoint" "slack_alerts" {
  name   = "Slack Alerts"
  url    = "https://hooks.slack.com/services/xxx"
  events = ["INVOICE_PAID", "ORGANIZATION_CREATED"]
}

# --- Template ---

resource "platformxe_templates_template" "welcome" {
  name    = "Welcome Email"
  subject = "Welcome to {{company_name}}"
  body_blocks_json = jsonencode([
    { type = "heading", content = "Welcome, {{name}}!", level = 1 },
    { type = "text", content = "Thanks for joining." },
  ])
}

# --- Workflow ---

resource "platformxe_workflows_trigger" "invoice_notify" {
  name           = "Invoice Payment Notification"
  trigger_config = jsonencode({ eventType = "INVOICE_PAID" })
  actions        = jsonencode([{ type = "webhook", webhookId = platformxe_webhooks_endpoint.slack_alerts.id }])
}

# --- Sending Domain ---

resource "platformxe_domains_sending" "main" {
  domain = "notifications.myapp.com"
}

# --- Event Subscription ---

resource "platformxe_events_subscription" "audit_events" {
  name        = "audit-events"
  event_types = ["ROLE_CREATED", "ROLE_UPDATED", "ROLE_DELETED"]
  webhook_url = "https://audit.myapp.com/events"
}

# --- Federation ---

resource "platformxe_permissions_federation_group" "caldera" {
  name = "Caldera Suite"
}

resource "platformxe_permissions_federation_member" "lettings" {
  group_id        = platformxe_permissions_federation_group.caldera.id
  organization_id = "org_lettings_123"
  prefix          = "LT"
}

resource "platformxe_permissions_federation_member" "chats" {
  group_id        = platformxe_permissions_federation_group.caldera.id
  organization_id = "org_chats_456"
  prefix          = "CH"
}

# --- Contextual Messaging (Threads) ---

resource "platformxe_threads_channel" "booking" {
  slug               = "booking"
  display_name       = "Booking Conversations"
  entity_type        = "BOOKING"
  participant_roles  = ["GUEST", "HOST", "PLATFORM"]
  default_visibility = ["ALL"]

  lifecycle_rules = jsonencode({
    autoClose = { onEntityStatus = ["CHECKED_OUT", "CANCELLED"] }
    autoArchive = { afterClosedDays = 90 }
    systemMessages = {
      onThreadCreated = "A new conversation has been started."
      onThreadClosed  = "This conversation has been closed."
    }
  })

  escalation_config = jsonencode({
    flagReasons = [
      { code = "SAFETY",  label = "Safety concern", severity = "HIGH" },
      { code = "DISPUTE", label = "Dispute",         severity = "MEDIUM" },
    ]
    autoDetection = [{
      patterns   = ["fire", "flooding", "gas leak", "smoke"]
      flagReason = "SAFETY"
      autoFlag   = true
    }]
    rules = [{
      id         = "rule-safety"
      name       = "Safety auto-escalation"
      trigger    = "PARTICIPANT_FLAG"
      conditions = { "in" = [{ "var" = "flag.reason" }, ["SAFETY"]] }
      actions    = [{
        type   = "CREATE_ISSUE"
        config = { title = "SAFETY: {{thread.subject}}", priority = "URGENT" }
      }]
      priority = 1
      isActive = true
    }]
  })
}

# --- Escalation Config Data Source ---

data "platformxe_threads_escalation_config" "booking" {
  channel_id = platformxe_threads_channel.booking.id
}

# --- Service Processors ---

resource "platformxe_ocr_processor" "main" {
  enabled              = true
  provider             = "azure"
  confidence_threshold = 0.90

  supported_document_types = ["passport", "drivers_license", "national_id"]
  languages                = ["en"]
}

resource "platformxe_pdf_processor" "main" {
  enabled          = true
  default_template = "standard"
  branding_source  = "organization"

  enabled_templates = ["standard", "offer-letter", "property-flyer"]
}

resource "platformxe_qr_processor" "main" {
  enabled        = true
  default_size   = 400
  default_format = "png"
  max_batch_size = 50
}

resource "platformxe_messaging_processor" "main" {
  enabled          = true
  email_enabled    = true
  email_from_name  = "MyApp Notifications"
  email_reply_to   = "support@myapp.com"

  sms_enabled        = true
  sms_default_region = "NG"

  whatsapp_enabled = false
}

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

resource "platformxe_exports_processor" "main" {
  enabled             = true
  allowed_formats     = ["csv", "json"]
  max_rows_per_export = 50000
  retention_days      = 14
}

resource "platformxe_identity_processor" "main" {
  enabled              = true
  verification_methods = ["bvn", "nin", "passport"]
  consent_required     = true
  ndpr_compliance      = true
}
