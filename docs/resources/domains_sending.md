---
page_title: "platformxe_domains_sending Resource"
description: "Manages a PlatformXe sending domain for email delivery."
---

# platformxe_domains_sending

Manages a sending domain for email delivery. After creation, configure the required DNS records and verify the domain.

## Example Usage

```hcl
resource "platformxe_domains_sending" "mail" {
  domain = "mail.example.com"
}

output "dns_records" {
  value = platformxe_domains_sending.mail.dns_records
}

output "first_record_type" {
  value = platformxe_domains_sending.mail.dns_records[0].type
}
```

## Argument Reference

- `domain` (String, Required) -- Domain name to configure for sending. Changing this forces replacement.

## Attribute Reference

- `id` (String) -- Domain ID, assigned by PlatformXe.
- `verified` (Boolean) -- Whether the domain DNS records have been verified.
- `dns_records` (List of Object) -- Required DNS records to configure for verification.
  - `type` (String) -- Record type (TXT, CNAME, MX).
  - `name` (String) -- Record name.
  - `value` (String) -- Record value.
  - `verified` (Boolean) -- Whether this individual record has been verified.

## Import

Sending domains can be imported using their ID:

```bash
terraform import platformxe_domains_sending.mail dom_abc123
```
