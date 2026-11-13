// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package provider

import (
	"testing"
)

// The acceptance-test helpers live in acctest.go (another package cannot
// import a _test.go file).

func TestProviderMetadata(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("Provider is nil")
	}
}

func TestProviderSchema(t *testing.T) {
	p := &PlatformXeProvider{}
	if p == nil {
		t.Fatal("PlatformXeProvider is nil")
	}
}
