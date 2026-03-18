# ═══════════════════════════════════════════════════════════════
# Standalone Floating IP Test
# ═══════════════════════════════════════════════════════════════
#
# This test creates a floating IP, verifies it, and can be destroyed.
#
# Usage:
#   cd tests/floating_ip_test
#   terraform init
#   terraform plan
#   terraform apply -auto-approve
#   terraform plan -detailed-exitcode   # idempotency check
#   terraform destroy -auto-approve
#
# ═══════════════════════════════════════════════════════════════

terraform {
  required_providers {
    acecloud = {
      source  = "registry.terraform.io/acecloud/acecloud"
      version = "0.1.0"
    }
  }
}

# ─── Variables ─────────────────────────────────────────────────

variable "api_url" {
  description = "AceCloud API base URL"
  type        = string
}

variable "email" {
  description = "Login email"
  type        = string
}

variable "password" {
  description = "Login password"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "AceCloud region"
  type        = string
}

variable "project_id" {
  description = "Cloud project UUID"
  type        = string
}

variable "external_network_id" {
  description = "External/public network UUID for floating IP allocation"
  type        = string
}

# ─── Provider ──────────────────────────────────────────────────

provider "acecloud" {
  api_url    = var.api_url
  email      = var.email
  password   = var.password
  region     = var.region
  project_id = var.project_id
}

# ─── Floating IP Resource ─────────────────────────────────────

resource "acecloud_floating_ip" "test" {
  floating_network_id = var.external_network_id
  billing_type        = "hourly"
  description         = "Standalone floating IP test"
}

# ─── Outputs ───────────────────────────────────────────────────

output "floating_ip_id" {
  description = "ID of the created floating IP"
  value       = acecloud_floating_ip.test.id
}

output "floating_ip_address" {
  description = "The allocated IP address"
  value       = acecloud_floating_ip.test.floating_ip_address
}

output "floating_ip_status" {
  description = "Status of the floating IP"
  value       = acecloud_floating_ip.test.status
}

output "floating_ip_billing_type" {
  description = "Billing type"
  value       = acecloud_floating_ip.test.billing_type
}
