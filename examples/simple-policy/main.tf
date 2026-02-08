# Simple Settings Catalog Policy Example
# This example shows the basic usage of creating a Settings Catalog policy

terraform {
  required_providers {
    intune = {
      source  = "tofutune/intune"
      version = "~> 0.1"
    }
  }
}

# Configure the Intune provider with Azure CLI authentication
provider "intune" {
  use_cli = true
  # Or use Service Principal:
  # tenant_id     = "your-tenant-id"
  # client_id     = "your-client-id"
  # client_secret = "your-client-secret"
}

# Create a simple Settings Catalog policy
resource "intune_settings_catalog_policy" "example" {
  name         = "Example Policy"
  description  = "A simple example policy"
  platforms    = "windows10AndLater"
  technologies = "mdm"
}

# Add some settings directly (without using a module)
resource "intune_settings_catalog_policy_settings" "example" {
  policy_id = intune_settings_catalog_policy.example.id

  # Disable real-time monitoring (for example purposes only!)
  setting {
    definition_id = "device_vendor_msft_defender_configuration_disablerealtimemonitoring"
    value_type    = "boolean"
    value         = "false"  # false = don't disable = enable real-time monitoring
  }

  # Enable cloud protection
  setting {
    definition_id = "device_vendor_msft_defender_configuration_allowcloudprotection"
    value_type    = "choice"
    value         = "device_vendor_msft_defender_configuration_allowcloudprotection_1"
  }
}

# Assign to all devices
resource "intune_policy_assignment" "example" {
  policy_id   = intune_settings_catalog_policy.example.id
  policy_type = "settings_catalog"

  all_devices = true
}

output "policy_id" {
  value = intune_settings_catalog_policy.example.id
}
