# Windows Update Module Outputs

output "policy_id" {
  description = "The ID of the policy these settings were added to"
  value       = var.policy_id
}

output "settings_configured" {
  description = "Summary of configured Windows Update settings"
  value = {
    automatic_update_mode        = var.automatic_update_mode
    feature_update_deferral_days = var.feature_update_deferral_days
    quality_update_deferral_days = var.quality_update_deferral_days
    active_hours                 = "${var.active_hours_start}:00 - ${var.active_hours_end}:00"
    feature_update_deadline_days = var.feature_update_deadline_days
    quality_update_deadline_days = var.quality_update_deadline_days
    grace_period_days            = var.grace_period_days
  }
}
