# Microsoft Defender Module Outputs

output "policy_id" {
  description = "The ID of the policy these settings were added to"
  value       = var.policy_id
}

output "settings_configured" {
  description = "Summary of configured Defender settings"
  value = {
    real_time_protection = var.real_time_protection
    behavior_monitoring  = var.behavior_monitoring
    cloud_protection     = var.cloud_protection
    cloud_block_level    = var.cloud_block_level
    pua_protection       = var.pua_protection
    network_protection   = var.network_protection
  }
}
