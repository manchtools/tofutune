# OneDrive Settings Module
# This module provides settings for OneDrive for Business configuration

terraform {
  required_providers {
    intune = {
      source = "MANCHTOOLS/tofutune"
    }
  }
}

resource "intune_settings_catalog_policy_settings" "onedrive" {
  policy_id = var.policy_id

  # ============================================================================
  # Known Folder Move (KFM) / Folder Backup
  # ============================================================================

  # Silently move Windows known folders to OneDrive
  dynamic "setting" {
    for_each = var.tenant_id != null && var.silently_move_known_folders ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_kfmsilentoptin"
      value_type    = "string"
      value         = var.tenant_id
    }
  }

  # Prompt users to move Windows known folders to OneDrive
  dynamic "setting" {
    for_each = var.tenant_id != null && var.prompt_to_move_known_folders ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_kfmoptinnowiznard"
      value_type    = "string"
      value         = var.tenant_id
    }
  }

  # Prevent users from redirecting known folders to their PC
  dynamic "setting" {
    for_each = var.prevent_redirect_to_pc != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_kfmblockoptout"
      value_type    = "choice"
      value         = var.prevent_redirect_to_pc ? "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_kfmblockoptout_1" : "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_kfmblockoptout_0"
    }
  }

  # ============================================================================
  # Sync Settings
  # ============================================================================

  # Silently sign in users to OneDrive sync
  dynamic "setting" {
    for_each = var.silent_account_config != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_silentaccountconfig"
      value_type    = "choice"
      value         = var.silent_account_config ? "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_silentaccountconfig_1" : "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_silentaccountconfig_0"
    }
  }

  # Use OneDrive Files On-Demand
  dynamic "setting" {
    for_each = var.files_on_demand_enabled != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_filesondemandsenabled"
      value_type    = "choice"
      value         = var.files_on_demand_enabled ? "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_filesondemandsenabled_1" : "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_filesondemandsenabled_0"
    }
  }

  # ============================================================================
  # Storage and Limits
  # ============================================================================

  # Set max size of user's OneDrive before download (GB)
  dynamic "setting" {
    for_each = var.max_size_before_download_gb != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_diskspacecheckthresholdmb"
      value_type    = "integer"
      value         = tostring(var.max_size_before_download_gb * 1024)  # Convert GB to MB
    }
  }

  # ============================================================================
  # Network Settings
  # ============================================================================

  # Limit download bandwidth (percentage)
  dynamic "setting" {
    for_each = var.download_bandwidth_limit_percent != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_automaticuploadbandwidthlimit"
      value_type    = "integer"
      value         = tostring(var.download_bandwidth_limit_percent)
    }
  }

  # Limit upload bandwidth (KB/s)
  dynamic "setting" {
    for_each = var.upload_bandwidth_limit_kbps != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_uploadbandwidthlimit"
      value_type    = "integer"
      value         = tostring(var.upload_bandwidth_limit_kbps)
    }
  }

  # ============================================================================
  # Sharing and Collaboration
  # ============================================================================

  # Prevent users from syncing personal OneDrive accounts
  dynamic "setting" {
    for_each = var.block_personal_account_sync != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_disablepersonalsync"
      value_type    = "choice"
      value         = var.block_personal_account_sync ? "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_disablepersonalsync_1" : "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_disablepersonalsync_0"
    }
  }

  # Allow syncing OneDrive accounts for only specific organizations
  dynamic "setting" {
    for_each = var.allowed_tenant_ids != null && length(var.allowed_tenant_ids) > 0 ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_allowtenantlist"
      value_type    = "collection"
      value         = jsonencode(var.allowed_tenant_ids)
    }
  }

  # Block syncing OneDrive accounts for specific organizations
  dynamic "setting" {
    for_each = var.blocked_tenant_ids != null && length(var.blocked_tenant_ids) > 0 ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_blocktenantlist"
      value_type    = "collection"
      value         = jsonencode(var.blocked_tenant_ids)
    }
  }

  # ============================================================================
  # Update Settings
  # ============================================================================

  # Set OneDrive update ring
  dynamic "setting" {
    for_each = var.update_ring != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_gpoupdatering"
      value_type    = "choice"
      value         = lookup({
        "enterprise"  = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_gpoupdatering_0"
        "production"  = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_gpoupdatering_4"
        "insiders"    = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_gpoupdatering_5"
        "deferred"    = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_gpoupdatering_64"
      }, var.update_ring, "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_gpoupdatering_4")
    }
  }

  # ============================================================================
  # User Experience
  # ============================================================================

  # Prevent OneDrive from generating network traffic until user signs in
  dynamic "setting" {
    for_each = var.prevent_network_traffic_until_signin != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_preventnetworktrafficuntilsignedin"
      value_type    = "choice"
      value         = var.prevent_network_traffic_until_signin ? "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_preventnetworktrafficuntilsignedin_1" : "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_preventnetworktrafficuntilsignedin_0"
    }
  }

  # Hide OneDrive shortcut on desktop
  dynamic "setting" {
    for_each = var.hide_desktop_shortcut != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_disabletutorial"
      value_type    = "choice"
      value         = var.hide_desktop_shortcut ? "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_disabletutorial_1" : "device_vendor_msft_policy_config_onedrivengscv2~policy~onedrivengsc_disabletutorial_0"
    }
  }
}
