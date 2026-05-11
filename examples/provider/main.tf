terraform {
  required_providers {
    acecloud = {
      source  = "registry.terraform.io/acecloud/acecloud"
      version = "~> 0.1"
    }
  }
  required_version = ">= 1.6"
}

# Authenticate with an API key (recommended for automation).
# Create one with: ace api-key create --service-name terraform-prod
provider "acecloud" {
  api_url              = var.acecloud_api_url
  api_key_id           = var.acecloud_api_key_id
  api_key_secret       = var.acecloud_api_key_secret
  api_key_service_name = "terraform"
  region               = var.acecloud_region
  project_id           = var.acecloud_project_id
}

# Alternative: authenticate with a JWT bearer token.
# Tokens expire after 24 hours; for long-running automation prefer api_key auth.
#
# provider "acecloud" {
#   api_url    = var.acecloud_api_url
#   api_token  = var.acecloud_api_token
#   region     = var.acecloud_region
#   project_id = var.acecloud_project_id
# }

variable "acecloud_api_url" {
  description = "Ace Cloud API base URL"
  type        = string
}

variable "acecloud_api_key_id" {
  description = "Ace Cloud API key identifier"
  type        = string
  sensitive   = true
}

variable "acecloud_api_key_secret" {
  description = "Ace Cloud API key secret"
  type        = string
  sensitive   = true
}

variable "acecloud_api_token" {
  description = "Ace Cloud API JWT token (alternative to API key)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "acecloud_region" {
  description = "Cloud region"
  type        = string
  default     = "mumbai"
}

variable "acecloud_project_id" {
  description = "AceCloud project UUID"
  type        = string
}
