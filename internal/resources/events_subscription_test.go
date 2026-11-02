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

func TestAccEventSubscription_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccSubscriptionConfig_basic(
					"https://example.com/events",
					[]string{"email.sent", "email.delivered"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_events_subscription.test", "id"),
					resource.TestCheckResourceAttrSet("platformxe_events_subscription.test", "secret"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "name", "tf-acc-subscription"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "webhook_url", "https://example.com/events"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "event_types.#", "2"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "event_types.0", "email.sent"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "event_types.1", "email.delivered"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "is_active", "true"),
				),
			},
		},
	})
}

func TestAccEventSubscription_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccSubscriptionConfig_basic(
					"https://example.com/events/v1",
					[]string{"email.sent"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_events_subscription.test", "id"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "event_types.#", "1"),
				),
			},
			// Update: change webhook URL and add event types
			{
				Config: testAccSubscriptionConfig_basic(
					"https://example.com/events/v2",
					[]string{"email.sent", "email.delivered", "email.bounced"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "webhook_url", "https://example.com/events/v2"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "event_types.#", "3"),
				),
			},
		},
	})
}

func TestAccEventSubscription_inactive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionConfig_inactive(
					"https://example.com/events/inactive",
					[]string{"email.sent"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_events_subscription.test", "id"),
					resource.TestCheckResourceAttr("platformxe_events_subscription.test", "is_active", "false"),
				),
			},
		},
	})
}

func testAccSubscriptionConfig_basic(webhookURL string, eventTypes []string) string {
	eventsHCL := ""
	for _, e := range eventTypes {
		eventsHCL += fmt.Sprintf("    %q,\n", e)
	}
	return fmt.Sprintf(`
%s

resource "platformxe_events_subscription" "test" {
  name        = "tf-acc-subscription"
  webhook_url = %q
  event_types = [
%s  ]
}
`, provider.AccProviderConfig(), webhookURL, eventsHCL)
}

func testAccSubscriptionConfig_inactive(webhookURL string, eventTypes []string) string {
	eventsHCL := ""
	for _, e := range eventTypes {
		eventsHCL += fmt.Sprintf("    %q,\n", e)
	}
	return fmt.Sprintf(`
%s

resource "platformxe_events_subscription" "test" {
  name        = "tf-acc-subscription"
  webhook_url = %q
  is_active   = false
  event_types = [
%s  ]
}
`, provider.AccProviderConfig(), webhookURL, eventsHCL)
}
