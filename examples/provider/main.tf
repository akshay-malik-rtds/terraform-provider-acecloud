terraform {
  required_providers {
    acecloud = {
      source  = "registry.terraform.io/akshay-malik-rtds/acecloud"
      version = "~> 0.2"
    }
  }
  required_version = ">= 1.6"
}

# Authenticate with an API key. Create one with:
#   ace api-key create --service-name terraform
provider "acecloud" {
  api_url              = var.acecloud_api_url
  api_key_id           = var.acecloud_api_key_id
  api_key_secret       = var.acecloud_api_key_secret
  api_key_service_name = "terraform"
  region               = var.acecloud_region
  project_id           = var.acecloud_project_id
}

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

variable "acecloud_region" {
  description = "Cloud region"
  type        = string
  default     = "mumbai"
}

variable "acecloud_project_id" {
  description = "AceCloud project UUID"
  type        = string
}
