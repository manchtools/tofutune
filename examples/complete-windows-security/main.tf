# Complete Windows Security Configuration Example
# This example demonstrates how to use the tofutune provider to configure
# a comprehensive Windows security baseline using modular settings.

terraform {
  required_providers {
    intune = {
      source  = "tofutune/intune"
      version = "~> 0.1"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.0"
    }
  }
}

# Configure the Intune provider
# Authentication will use the same credentials as the azuread provider
provider "intune" {
  # Uses ARM_TENANT_ID, ARM_CLIENT_ID, ARM_CLIENT_SECRET from environment
  # Or use Azure CLI authentication by running 'az login' first
}

# Get Azure AD groups for assignments
data "azuread_group" "all_devices" {
  display_name = "All Devices"
}

data "azuread_group" "test_devices" {
  display_name = "Test Devices"
}

# ============================================================================
# Settings Catalog Policy with Modular Settings
# ============================================================================

# Create the base policy with inline assignment
resource "intune_settings_catalog_policy" "windows_security" {
  name         = "Windows Security Baseline"
  description  = "Comprehensive Windows security configuration"
  platforms    = "windows10AndLater"
  technologies = "mdm"

  # Assign to all devices, excluding test devices
  assignment {
    include_groups = [data.azuread_group.all_devices.id]
    exclude_groups = [data.azuread_group.test_devices.id]
  }
}

# Add Microsoft Defender settings using the pre-built module
module "defender" {
  source = "../../modules/defender"

  policy_id = intune_settings_catalog_policy.windows_security.id

  # Enable all protection features
  real_time_protection = true
  behavior_monitoring  = true
  cloud_protection     = "advanced"
  cloud_block_level    = "high"
  pua_protection       = "enabled"
  script_scanning      = true
  archive_scanning     = true
  network_protection   = "enabled"
  intrusion_prevention = true

  # Schedule a weekly full scan on Sundays at 2 AM
  scheduled_scan_day  = 1  # Sunday
  scheduled_scan_time = 120  # 2:00 AM
  scheduled_scan_type = "full"
}

# Add BitLocker settings using the pre-built module
module "bitlocker" {
  source = "../../modules/bitlocker"

  policy_id = intune_settings_catalog_policy.windows_security.id

  require_device_encryption     = true
  os_drive_encryption_method    = "xts_aes256"
  fixed_drive_encryption_method = "xts_aes256"
  require_tpm                   = true
  recovery_password_rotation    = "azure_ad_only"
  backup_recovery_to_azure_ad   = true
  hide_recovery_options         = true
}

# Add Windows Update settings using the pre-built module
module "windows_update" {
  source = "../../modules/windows_update"

  policy_id = intune_settings_catalog_policy.windows_security.id

  automatic_update_mode        = "auto_install_and_reboot_at_scheduled_time"
  feature_update_deferral_days = 14   # Wait 14 days for feature updates
  quality_update_deferral_days = 3    # Wait 3 days for quality updates
  active_hours_start           = 8    # 8 AM
  active_hours_end             = 18   # 6 PM
  feature_update_deadline_days = 7
  quality_update_deadline_days = 3
  grace_period_days            = 2
  allow_driver_updates         = true
}

# ============================================================================
# Compliance Policy
# ============================================================================

resource "intune_compliance_policy" "windows" {
  display_name = "Windows 10 Compliance Policy"
  description  = "Ensures devices meet security requirements"

  # Password requirements
  password_required       = true
  password_minimum_length = 8
  password_required_type  = "alphanumeric"
  password_block_simple   = true

  # Security requirements
  bitlocker_enabled      = true
  secure_boot_enabled    = true
  code_integrity_enabled = true
  tpm_required           = true

  # Defender requirements
  defender_enabled      = true
  rtp_enabled           = true
  antivirus_required    = true
  anti_spyware_required = true

  # Firewall
  active_firewall_required = true

  # OS version requirements
  os_minimum_version = "10.0.19041.0"

  # Actions for non-compliant devices
  scheduled_actions_for_rule {
    rule_name = "DeviceNotCompliant"
    scheduled_action_configurations {
      action_type        = "block"
      grace_period_hours = 24  # Give users 24 hours to become compliant
    }
  }

  # Assign to all devices
  assignment {
    all_devices = true
  }
}

# ============================================================================
# Outputs
# ============================================================================

output "security_policy_id" {
  description = "The ID of the Windows Security Settings Catalog policy"
  value       = intune_settings_catalog_policy.windows_security.id
}

output "compliance_policy_id" {
  description = "The ID of the Windows Compliance policy"
  value       = intune_compliance_policy.windows.id
}

output "defender_settings" {
  description = "Summary of Defender settings applied"
  value       = module.defender.settings_configured
}

output "bitlocker_settings" {
  description = "Summary of BitLocker settings applied"
  value       = module.bitlocker.settings_configured
}

output "windows_update_settings" {
  description = "Summary of Windows Update settings applied"
  value       = module.windows_update.settings_configured
}
