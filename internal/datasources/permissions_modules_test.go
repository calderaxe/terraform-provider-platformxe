// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/calderaxe/terraform-provider-platformxe/internal/provider"
)

func TestAccDataSourceModules_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceModulesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The modules list should be populated (at least zero items is valid)
					resource.TestCheckResourceAttrSet("data.platformxe_permissions_modules.test", "modules.#"),
				),
			},
		},
	})
}

func TestAccDataSourceModules_hasEntries(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceModulesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the first module has expected attributes when modules exist
					resource.TestCheckResourceAttrSet("data.platformxe_permissions_modules.test", "modules.0.id"),
					resource.TestCheckResourceAttrSet("data.platformxe_permissions_modules.test", "modules.0.app"),
					resource.TestCheckResourceAttrSet("data.platformxe_permissions_modules.test", "modules.0.name"),
				),
			},
		},
	})
}

func testAccDataSourceModulesConfig() string {
	return fmt.Sprintf(`
%s

data "platformxe_permissions_modules" "test" {}
`, provider.AccProviderConfig())
}
