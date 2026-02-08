# Microsoft Edge Module Outputs

output "policy_id" {
  description = "The ID of the policy these settings were added to"
  value       = var.policy_id
}

output "settings_configured" {
  description = "Summary of configured Edge settings"
  value = {
    startup = {
      homepage_url   = var.homepage_url
      startup_action = var.startup_action
    }
    security = {
      smartscreen          = var.enable_smartscreen
      smartscreen_pua      = var.smartscreen_pua_enabled
      site_isolation       = var.site_isolation_enabled
    }
    privacy = {
      tracking_prevention = var.tracking_prevention_mode
      do_not_track        = var.do_not_track_enabled
    }
    features = {
      password_manager = var.password_manager_enabled
      copilot_sidebar  = !coalesce(var.disable_copilot_sidebar, false)
      sync             = !coalesce(var.sync_disabled, false)
    }
  }
}
