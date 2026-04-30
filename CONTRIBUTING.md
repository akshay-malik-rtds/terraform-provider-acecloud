# Contributing to the AceCloud Terraform Provider

Thank you for your interest in contributing! This document outlines how to set up your development environment, propose changes, and submit pull requests.

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](./CODE_OF_CONDUCT.md). By participating you agree to abide by its terms.

## Reporting Issues

- **Bugs**: Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md) template
- **Features**: Use the [Feature Request](.github/ISSUE_TEMPLATE/feature_request.md) template
- **Security**: Do **not** open a public issue. See [SECURITY.md](./SECURITY.md)

When reporting a bug, please include:
- Provider version
- Terraform version (`terraform version`)
- Minimal Terraform configuration that reproduces the issue
- Expected vs actual behavior
- Output of `terraform apply` with `TF_LOG=DEBUG`

## Development Setup

### Prerequisites

- Go 1.22 or later (`go version`)
- Terraform 1.6 or later
- An [AceCloud](https://acecloud.ai) account with API access (for acceptance testing)

### Building

```bash
git clone https://github.com/akshay-malik-rtds/terraform-provider-acecloud
cd terraform-provider-acecloud
make build
```

### Local Provider Testing

To test your local build with Terraform configurations, use [`dev_overrides`](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers):

1. Build and install:
   ```bash
   make install
   ```

2. Add to `~/.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "akshay-malik-rtds/acecloud" = "/path/to/terraform-provider-acecloud"
     }
     direct {}
   }
   ```

3. Run your Terraform configuration normally — it will use your local build.

### Running Tests

**Unit tests** (no API access required):

```bash
make test
```

**Acceptance tests** (require real AceCloud credentials and **will create real resources**):

```bash
export ACECLOUD_API_URL="https://api.acecloud.ai/api/v1"
export ACECLOUD_API_TOKEN="your-token"
export ACECLOUD_REGION="ap-south-noi-1"
export ACECLOUD_PROJECT_ID="your-project-uuid"
make testacc
```

### Documentation

Provider docs are auto-generated from schema descriptions and example files. After modifying schemas or adding examples, regenerate docs:

```bash
make docs
```

Validate the generated docs match the registry's expected format:

```bash
make docs-validate
```

## Pull Request Workflow

1. **Open an issue first** for non-trivial changes so we can discuss the approach.
2. **Fork the repo** and create a feature branch from `main`:
   ```bash
   git checkout -b feature/my-change
   ```
3. **Make your changes**:
   - Follow the existing code style (Go's standard formatting + the patterns used elsewhere in the codebase)
   - Add unit tests for new functionality
   - Update relevant docs
   - Run `make fmt`, `make lint`, and `make test` before committing
4. **Commit messages** should follow [Conventional Commits](https://www.conventionalcommits.org/) where practical:
   - `feat(instance): add support for spot pricing`
   - `fix(volume): correct resize behavior on async API`
   - `docs: clarify VPC subnet inline requirements`
5. **Open a PR** referencing the issue. Use the PR template.
6. **CI must pass**: build, tests, lint, and docs validation all run on every PR.

## Code Style

- **Errors**: Use the format `Failed to <verb> <resource>` (e.g. `"Failed to create instance"`).
- **Schema descriptions**: User-facing — avoid internal infrastructure terminology.
- **Wait/retry**: Use the `internal/wait` package, never hand-rolled loops.
- **API calls**: Always go through `internal/client`. Never construct backend service URLs directly.

## Resource Coding Standards

Detailed coding standards live in the project's `CLAUDE.md` (developer-facing). Key rules:

1. Error messages: `"Failed to <verb> <resource>"`
2. Computed fields: always set to known values after API calls
3. Unit tests: minimum 25 tests per resource
4. Schema validation matches the AceCloud platform validation rules

## Releasing

Releases are tagged from `main` using semver:

```bash
git tag v0.X.Y
git push origin v0.X.Y
```

The `release` GitHub Action handles the rest (multi-platform builds, GPG signing, GitHub release, registry publication).

## Questions?

- Open a [discussion](https://github.com/akshay-malik-rtds/terraform-provider-acecloud/discussions) for general questions
- Check the [provider documentation](https://registry.terraform.io/providers/akshay-malik-rtds/acecloud/latest/docs) for usage help

Thanks for contributing!
