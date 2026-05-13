// Package tools tracks development tools as Go module dependencies so
// they're versioned alongside the provider source. The build tag ensures
// these imports don't end up in the production binary.
//
// Usage: `go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs ...`
//
//go:build tools

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
