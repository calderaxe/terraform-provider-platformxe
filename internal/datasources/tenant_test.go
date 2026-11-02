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

func TestAccDataSourceTenant_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.AccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTenantConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.platformxe_tenant.test", "id"),
					resource.TestCheckResourceAttrSet("data.platformxe_tenant.test", "name"),
					resource.TestCheckResourceAttrSet("data.platformxe_tenant.test", "slug"),
					resource.TestCheckResourceAttrSet("data.platformxe_tenant.test", "plan"),
					resource.TestCheckResourceAttrSet("data.platformxe_tenant.test", "region"),
					resource.TestCheckResourceAttr("data.platformxe_tenant.test", "is_active", "true"),
				),
			},
		},
	})
}

func testAccDataSourceTenantConfig() string {
	return fmt.Sprintf(`
%s

data "platformxe_tenant" "test" {}
`, provider.AccProviderConfig())
}
