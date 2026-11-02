---
page_title: "platformxe_templates_template Resource"
description: "Manages a PlatformXe content template."
---

# platformxe_templates_template

Manages a reusable content template for email and other messaging. The body is a list of typed blocks that PlatformXe renders; `{{variable}}` placeholders are filled at render time.

## Example Usage

```hcl
resource "platformxe_templates_template" "welcome" {
  name    = "welcome-email"
  subject = "Welcome to {{company}}"

  body_blocks_json = jsonencode([
    { type = "heading", content = "Welcome, {{name}}!", level = 1 },
    { type = "text", content = "Thanks for joining {{company}}." },
    { type = "button", label = "Open your dashboard", url = "https://app.example.com" },
  ])
}
```

## Argument Reference

- `name` (String, Required) -- Template name, unique within the organization.
- `subject` (String, Required) -- Subject line. Supports `{{variable}}` interpolation.
- `body_blocks_json` (String, Required) -- JSON array of body blocks (use `jsonencode`). Each block has a `type` -- `text`, `heading`, `button`, `image`, `divider` or `spacer` -- and the fields that type uses: `content`, `level`, `label`, `url`, `src`, `alt`, `height`. A change made outside Terraform shows as a diff; formatting differences do not.

## Attribute Reference

- `id` (String) -- Template ID, assigned by PlatformXe.

## Import

Templates can be imported using their ID:

```bash
terraform import platformxe_templates_template.welcome tmpl_abc123
```
