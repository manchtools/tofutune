# Windows Privacy Module Outputs

output "policy_id" {
  description = "The ID of the policy these settings were added to"
  value       = var.policy_id
}

output "settings_configured" {
  description = "Summary of configured privacy settings"
  value = {
    telemetry = {
      level                         = var.telemetry_level
      limit_diagnostic_log          = var.limit_diagnostic_log_collection
      limit_dump_collection         = var.limit_dump_collection
      device_name_in_telemetry      = var.allow_device_name_in_telemetry
    }
    personalization = {
      tailored_experiences = var.allow_tailored_experiences
      advertising_id       = !var.disable_advertising_id
    }
    permissions = {
      location   = var.allow_location
      camera     = var.allow_camera
      microphone = var.allow_microphone
    }
    activity = {
      activity_feed     = var.enable_activity_feed
      publish_activities = var.publish_user_activities
      upload_activities  = var.upload_user_activities
    }
  }
}
