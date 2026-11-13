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

func TestAccThreadChannel_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccThreadChannelConfig_basic(
					"tf-acc-booking",
					"Booking Conversations",
					"BOOKING",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_threads_channel.test", "id"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "slug", "tf-acc-booking"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "display_name", "Booking Conversations"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "entity_type", "BOOKING"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "participant_roles.#", "3"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "default_visibility.#", "1"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "default_visibility.0", "ALL"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "is_active", "true"),
				),
			},
		},
	})
}

func TestAccThreadChannel_withLifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccThreadChannelConfig_withLifecycle(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_threads_channel.test", "id"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "slug", "tf-acc-service-order"),
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "entity_type", "SERVICE_ORDER"),
					resource.TestCheckResourceAttrSet("platformxe_threads_channel.test", "lifecycle_rules"),
				),
			},
		},
	})
}

func TestAccThreadChannel_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccThreadChannelConfig_basic(
					"tf-acc-update-ch",
					"Original Name",
					"APPOINTMENT",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "display_name", "Original Name"),
				),
			},
			// Update display_name (slug and entity_type are ForceNew so cannot be changed)
			{
				Config: testAccThreadChannelConfig_basic(
					"tf-acc-update-ch",
					"Updated Channel Name",
					"APPOINTMENT",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("platformxe_threads_channel.test", "display_name", "Updated Channel Name"),
				),
			},
		},
	})
}

func testAccThreadChannelConfig_basic(slug, displayName, entityType string) string {
	return fmt.Sprintf(`
%s

resource "platformxe_threads_channel" "test" {
  slug         = %q
  display_name = %q
  entity_type  = %q

  participant_roles  = ["GUEST", "HOST", "PLATFORM"]
  default_visibility = ["ALL"]
}
`, provider.AccProviderConfig(), slug, displayName, entityType)
}

func testAccThreadChannelConfig_withLifecycle() string {
	return fmt.Sprintf(`
%s

resource "platformxe_threads_channel" "test" {
  slug         = "tf-acc-service-order"
  display_name = "Service Order Threads"
  entity_type  = "SERVICE_ORDER"

  participant_roles  = ["CUSTOMER", "TECHNICIAN", "PLATFORM"]
  default_visibility = ["ALL"]

  lifecycle_rules = jsonencode({
    autoCloseOnStatus  = ["COMPLETED", "CANCELLED"]
    autoArchiveDays    = 30
    retentionDays      = 90
    systemMessages     = true
  })
}
`, provider.AccProviderConfig())
}
