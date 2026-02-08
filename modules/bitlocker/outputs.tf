# BitLocker Module Outputs

output "policy_id" {
  description = "The ID of the policy these settings were added to"
  value       = var.policy_id
}

output "settings_configured" {
  description = "Summary of configured BitLocker settings"
  value = {
    require_device_encryption         = var.require_device_encryption
    os_drive_encryption_method        = var.os_drive_encryption_method
    fixed_drive_encryption_method     = var.fixed_drive_encryption_method
    removable_drive_encryption_method = var.removable_drive_encryption_method
    require_tpm                       = var.require_tpm
    backup_recovery_to_azure_ad       = var.backup_recovery_to_azure_ad
  }
}
