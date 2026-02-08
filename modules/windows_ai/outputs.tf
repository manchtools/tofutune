# Windows AI Module Outputs

output "policy_id" {
  description = "The ID of the policy these settings were added to"
  value       = var.policy_id
}

output "settings_configured" {
  description = "Summary of configured Windows AI settings"
  value = {
    copilot = {
      disabled_user   = var.disable_copilot
      disabled_device = var.disable_copilot_device
      allowed         = var.allow_copilot
    }
    recall = {
      disabled           = var.disable_recall
      snapshots_disabled = var.disable_recall_snapshots
    }
    generative_ai = {
      image_creator_disabled = var.disable_image_creator
      cocreator_disabled     = var.disable_cocreator
    }
    search = {
      highlights_disabled   = var.disable_search_highlights
      cloud_search_disabled = var.disable_cloud_search
    }
    suggestions = {
      tips_disabled       = var.disable_tips_notifications
      spotlight_disabled  = var.disable_spotlight
      settings_suggestions_disabled = var.disable_settings_suggestions
    }
  }
}
