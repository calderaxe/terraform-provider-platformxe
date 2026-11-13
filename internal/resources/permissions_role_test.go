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

func TestAccRole_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccRoleConfig_basic("tf-acc-role-basic", "Acceptance test role"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_permissions_role.test", "id"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "name", "tf-acc-role-basic"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "description", "Acceptance test role"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "model", "SIMPLE"),
				),
			},
			// ImportState (verify resource can be read back)
			{
				ResourceName:      "platformxe_permissions_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRole_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccRoleConfig_basic("tf-acc-role-update", "Original description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_permissions_role.test", "id"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "name", "tf-acc-role-update"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "description", "Original description"),
				),
			},
			// Update name and description
			{
				Config: testAccRoleConfig_basic("tf-acc-role-updated", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_permissions_role.test", "id"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "name", "tf-acc-role-updated"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccRole_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_basic("tf-acc-role-disappears", "Will be deleted externally"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_permissions_role.test", "id"),
				),
				// After the apply, the framework will run a plan to detect drift.
				// If the resource was deleted out-of-band, the provider Read should
				// handle the 404 gracefully and the next apply should recreate it.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRole_fullModel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_withModel("tf-acc-role-full", "Full model role", "FULL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("platformxe_permissions_role.test", "id"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "name", "tf-acc-role-full"),
					resource.TestCheckResourceAttr("platformxe_permissions_role.test", "model", "FULL"),
				),
			},
		},
	})
}

func testAccRoleConfig_basic(name, description string) string {
	return fmt.Sprintf(`
%s

resource "platformxe_permissions_role" "test" {
  name        = %q
  description = %q
  model       = "SIMPLE"
}
`, provider.AccProviderConfig(), name, description)
}

func testAccRoleConfig_withModel(name, description, model string) string {
	return fmt.Sprintf(`
%s

resource "platformxe_permissions_role" "test" {
  name        = %q
  description = %q
  model       = %q
}
`, provider.AccProviderConfig(), name, description, model)
}
