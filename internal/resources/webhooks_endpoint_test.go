// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/calderaxe/terraform-provider-platformxe/internal/provider"
)

func TestAccWebhook_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWebhookConfig_basic(
					"tf-acc-webhook",
					"https://example.com/webhook",
					[]string{"email.sent", "email.delivered"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_webhooks_endpoint.test", "id"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "name", "tf-acc-webhook"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "events.#", "2"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "events.0", "email.sent"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "events.1", "email.delivered"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "is_active", "true"),
					resource.TestCheckResourceAttrSet("platformxe_webhooks_endpoint.test", "secret"),
				),
			},
		},
	})
}

func TestAccWebhook_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccWebhookConfig_basic(
					"tf-acc-webhook-update",
					"https://example.com/webhook/v1",
					[]string{"email.sent"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_webhooks_endpoint.test", "id"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "events.#", "1"),
				),
			},
			// Update: add more events and change URL
			{
				Config: testAccWebhookConfig_basic(
					"tf-acc-webhook-update",
					"https://example.com/webhook/v2",
					[]string{"email.sent", "email.delivered", "email.bounced"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_webhooks_endpoint.test", "id"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "url", "https://example.com/webhook/v2"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "events.#", "3"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "events.0", "email.sent"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "events.1", "email.delivered"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "events.2", "email.bounced"),
				),
			},
		},
	})
}

func TestAccWebhook_inactive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_inactive(
					"tf-acc-webhook-inactive",
					"https://example.com/webhook/inactive",
					[]string{"email.sent"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_webhooks_endpoint.test", "id"),
					resource.TestCheckResourceAttr("platformxe_webhooks_endpoint.test", "is_active", "false"),
				),
			},
		},
	})
}

func testAccWebhookConfig_basic(name, url string, events []string) string {
	eventsHCL := ""
	for _, e := range events {
		eventsHCL += fmt.Sprintf("    %q,\n", e)
	}
	return fmt.Sprintf(`
%s

resource "platformxe_webhooks_endpoint" "test" {
  name   = %q
  url    = %q
  events = [
%s  ]
}
`, provider.AccProviderConfig(), name, url, eventsHCL)
}

func testAccWebhookConfig_inactive(name, url string, events []string) string {
	eventsHCL := ""
	for _, e := range events {
		eventsHCL += fmt.Sprintf("    %q,\n", e)
	}
	return fmt.Sprintf(`
%s

resource "platformxe_webhooks_endpoint" "test" {
  name      = %q
  url       = %q
  is_active = false
  events    = [
%s  ]
}
`, provider.AccProviderConfig(), name, url, eventsHCL)
}
