# API key for a CI/CD pipeline. The secret is returned only on create
# and stored in Terraform state thereafter — save it to a secret manager
# immediately, since it cannot be recovered later.
#
# Note: new API keys must be created from the AceCloud console. This
# resource manages an existing key's metadata, updates, and deletion
# via `terraform import`.

resource "acecloud_api_key" "ci" {
  service_name = "ci-pipeline"
  description  = "Long-lived credential used by GitHub Actions to run terraform"
}

# Pass the credential parts to your secret manager / CI environment.
output "ci_api_key_id" {
  description = "AceCloud API key ID for the CI pipeline."
  value       = acecloud_api_key.ci.id
  sensitive   = true
}

output "ci_api_key_secret" {
  description = "AceCloud API key secret. Use together with the ID as 'X-Ace-Api-Key: <id>.<secret>'."
  value       = acecloud_api_key.ci.secret
  sensitive   = true
}

# Disable a key without deleting it (e.g. during incident response).
resource "acecloud_api_key" "rotation_target" {
  service_name = "legacy-script"
  description  = "Pinned to disabled until rotated"
  enabled      = false
}
