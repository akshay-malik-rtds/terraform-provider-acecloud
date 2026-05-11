// Package acctest provides shared helpers for Terraform acceptance tests.
//
// Acceptance tests run against a real Ace Cloud backend and are gated by the
// TF_ACC=1 environment variable (the standard Terraform Plugin convention).
// They are skipped automatically during normal `go test ./...` runs.
//
// Required environment variables when TF_ACC=1:
//
//	ACECLOUD_API_URL              base URL, e.g. https://dev-portal.acecloud.ai/api/v1
//	ACECLOUD_API_KEY_ID           IAM API key id (uuid)
//	ACECLOUD_API_KEY_SECRET       IAM API key secret
//	ACECLOUD_API_KEY_SERVICE_NAME service name bound to the key
//	ACECLOUD_REGION               region, e.g. ap-south-noi-1
//	ACECLOUD_PROJECT_ID           project id (32 hex chars, no dashes)
//
// Optional helper environment variables (used by individual tests):
//
//	ACECLOUD_FLAVOR_ID            compute flavor uuid
//	ACECLOUD_IMAGE_ID             image uuid
//	ACECLOUD_EXTERNAL_NETWORK_ID  external network uuid for FIPs / routers
package acctest

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/provider"
)

// ProviderName is the address used in `required_providers` blocks of test
// configs. Acceptance tests use this single name to refer to the in-process
// provider factory.
const ProviderName = "acecloud"

// envVars lists the environment variables required for acceptance tests.
// PreCheck verifies all of these are set before any test runs.
var envVars = []string{
	"ACECLOUD_API_URL",
	"ACECLOUD_API_KEY_ID",
	"ACECLOUD_API_KEY_SECRET",
	"ACECLOUD_API_KEY_SERVICE_NAME",
	"ACECLOUD_REGION",
	"ACECLOUD_PROJECT_ID",
}

// PreCheck is called by every acceptance test to verify the environment is
// configured. It fails the test fast (before any apply) if a required
// variable is missing.
func PreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped (set TF_ACC=1 to enable)")
	}
	for _, v := range envVars {
		if os.Getenv(v) == "" {
			t.Fatalf("acceptance test requires %s to be set", v)
		}
	}
}

// PreCheckOptional skips a test unless the named env var is set. Used by
// tests that need extra inputs like a flavor or image UUID.
func PreCheckOptional(t *testing.T, name string) {
	t.Helper()
	if os.Getenv(name) == "" {
		t.Skipf("test requires %s to be set", name)
	}
}

// EnvOrDefault returns the value of envVar if set, else the default.
func EnvOrDefault(envVar, def string) string {
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	return def
}

// providerVersion is the version reported by the in-process provider during
// acceptance tests. This value is arbitrary; it does not influence behavior.
const providerVersion = "test"

// providerFactoriesOnce ensures the factory map is built only once per
// process. Tests share the same in-process provider instance via the
// terraform-plugin-testing helper.
var (
	providerFactoriesOnce sync.Once
	providerFactoriesVal  map[string]func() (tfprotov6.ProviderServer, error)
)

// ProtoV6ProviderFactories returns the factory map used by acceptance tests.
// Each test wires this into resource.TestCase.ProtoV6ProviderFactories.
func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	providerFactoriesOnce.Do(func() {
		providerFactoriesVal = map[string]func() (tfprotov6.ProviderServer, error){
			ProviderName: providerserver.NewProtocol6WithError(provider.New(providerVersion)()),
		}
	})
	return providerFactoriesVal
}

// ProviderConfig returns an HCL `provider` block that all test configs
// prepend. With ProtoV6ProviderFactories the provider source is registered
// under the local factory name `acecloud`, so test configs do not need an
// explicit `terraform { required_providers { ... } }` block.
//
// Credentials are pulled from the test environment so secrets never appear
// in literal HCL.
func ProviderConfig() string {
	return fmt.Sprintf(`
provider "acecloud" {
  api_url              = %q
  api_key_id           = %q
  api_key_secret       = %q
  api_key_service_name = %q
  region               = %q
  project_id           = %q
}
`,
		os.Getenv("ACECLOUD_API_URL"),
		os.Getenv("ACECLOUD_API_KEY_ID"),
		os.Getenv("ACECLOUD_API_KEY_SECRET"),
		os.Getenv("ACECLOUD_API_KEY_SERVICE_NAME"),
		os.Getenv("ACECLOUD_REGION"),
		os.Getenv("ACECLOUD_PROJECT_ID"),
	)
}

// RandomName returns a unique name short enough to fit common 32-char
// limits even when callers append a few extra characters in their HCL configs.
// Format: "tfa-<prefix>-<8 hex chars>" — about 13 + len(prefix) characters.
func RandomName(prefix string) string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand from the OS does not realistically fail; if it does,
		// fall back to a deterministic-but-still-unique value derived from
		// the prefix length so tests can still run.
		return fmt.Sprintf("tfa-%s-%08x", prefix, len(prefix))
	}
	return fmt.Sprintf("tfa-%s-%s", prefix, hex.EncodeToString(b[:]))
}

// FlavorID returns the flavor UUID from ACECLOUD_FLAVOR_ID, skipping the
// test when unset.
func FlavorID(t *testing.T) string {
	t.Helper()
	PreCheckOptional(t, "ACECLOUD_FLAVOR_ID")
	return os.Getenv("ACECLOUD_FLAVOR_ID")
}

// ImageID returns the image UUID from ACECLOUD_IMAGE_ID, skipping when unset.
func ImageID(t *testing.T) string {
	t.Helper()
	PreCheckOptional(t, "ACECLOUD_IMAGE_ID")
	return os.Getenv("ACECLOUD_IMAGE_ID")
}

// ExternalNetworkID returns the external network UUID, skipping when unset.
func ExternalNetworkID(t *testing.T) string {
	t.Helper()
	PreCheckOptional(t, "ACECLOUD_EXTERNAL_NETWORK_ID")
	return os.Getenv("ACECLOUD_EXTERNAL_NETWORK_ID")
}
