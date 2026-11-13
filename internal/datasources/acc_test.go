// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package datasources_test

import (
	"os"
	"testing"
)

// testAccPreCheck validates that required environment variables are set before
// running acceptance tests. All acceptance tests call this in their PreCheck.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC is set")
	}

	if os.Getenv("PLATFORMXE_API_KEY") == "" {
		t.Fatal("PLATFORMXE_API_KEY must be set for acceptance tests")
	}
}
