# OneDrive Module Outputs

output "policy_id" {
  description = "The ID of the policy these settings were added to"
  value       = var.policy_id
}

output "settings_configured" {
  description = "Summary of configured OneDrive settings"
  value = {
    known_folder_move = {
      tenant_id               = var.tenant_id
      silent_move             = var.silently_move_known_folders
      prompt_move             = var.prompt_to_move_known_folders
      prevent_redirect_to_pc  = var.prevent_redirect_to_pc
    }
    sync = {
      silent_account_config = var.silent_account_config
      files_on_demand       = var.files_on_demand_enabled
    }
    restrictions = {
      block_personal_accounts = var.block_personal_account_sync
      allowed_tenants         = var.allowed_tenant_ids
      blocked_tenants         = var.blocked_tenant_ids
    }
    network = {
      download_limit_percent = var.download_bandwidth_limit_percent
      upload_limit_kbps      = var.upload_bandwidth_limit_kbps
    }
    update_ring = var.update_ring
  }
}
