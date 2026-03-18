terraform {
  required_providers {
    acecloud = {
      source  = "registry.terraform.io/acecloud/acecloud"
      version = "~> 0.1"
    }
  }
  required_version = ">= 1.6"
}

provider "acecloud" {
  api_url    = var.acecloud_api_url
  api_token  = var.acecloud_api_token
  region     = var.acecloud_region
  project_id = var.acecloud_project_id
}

variable "acecloud_api_url" {
  description = "Ace Cloud API base URL"
  type        = string
}

variable "acecloud_api_token" {
  description = "Ace Cloud API JWT token"
  type        = string
  sensitive   = true
}

variable "acecloud_region" {
  description = "Cloud region"
  type        = string
  default     = "mumbai"
}

variable "acecloud_project_id" {
  description = "OpenStack project UUID"
  type        = string
}
