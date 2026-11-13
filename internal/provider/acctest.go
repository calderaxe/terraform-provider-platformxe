// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG:
// 2026-11-01  Moved out of provider_test.go. The resources_test and
//             datasources_test packages import these helpers, and a _test.go
//             file is not importable from another package, so the acceptance
//             tests never compiled; a TestAcc… function in a _test.go file
//             must also take *testing.T (go vet). Renamed without the Test
//             prefix so they are not mistaken for tests.
// =============================================================================

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// AccProtoV6ProviderFactories returns the provider factories that the
// acceptance tests of the resource and data source packages run against
// (Protocol V6, as the Terraform Plugin Framework expects).
func AccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"platformxe": providerserver.NewProtocol6WithError(New()),
	}
}

// AccProviderConfig returns the provider block for acceptance-test
// configurations. The API key and optional base URL come from the
// environment (PLATFORMXE_API_KEY, PLATFORMXE_BASE_URL), never from a file.
func AccProviderConfig() string {
	return `
provider "platformxe" {}
`
}
