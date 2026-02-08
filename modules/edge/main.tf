# Microsoft Edge Browser Settings Module
# This module provides settings for Microsoft Edge browser configuration

terraform {
  required_providers {
    intune = {
      source = "MANCHTOOLS/tofutune"
    }
  }
}

resource "intune_settings_catalog_policy_settings" "edge" {
  policy_id = var.policy_id

  # ============================================================================
  # Homepage and Startup
  # ============================================================================

  # Homepage URL
  dynamic "setting" {
    for_each = var.homepage_url != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~startup_homepageisnewtabpage"
      value_type    = "choice"
      value         = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~startup_homepageisnewtabpage_0"
    }
  }

  dynamic "setting" {
    for_each = var.homepage_url != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~startup_homepagelocation"
      value_type    = "string"
      value         = var.homepage_url
    }
  }

  # Startup action (0=Open home, 1=Restore last session, 5=Open URLs)
  dynamic "setting" {
    for_each = var.startup_action != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~startup_restoreonstartup"
      value_type    = "choice"
      value         = lookup({
        "open_new_tab"      = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~startup_restoreonstartup_5"
        "restore_session"   = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~startup_restoreonstartup_1"
        "open_urls"         = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~startup_restoreonstartup_4"
      }, var.startup_action, "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~startup_restoreonstartup_5")
    }
  }

  # ============================================================================
  # Security Settings
  # ============================================================================

  # SmartScreen for Microsoft Edge
  dynamic "setting" {
    for_each = var.enable_smartscreen != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_smartscreenenabled"
      value_type    = "choice"
      value         = var.enable_smartscreen ? "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_smartscreenenabled_1" : "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_smartscreenenabled_0"
    }
  }

  # SmartScreen for PUA (Potentially Unwanted Apps)
  dynamic "setting" {
    for_each = var.smartscreen_pua_enabled != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_smartscreenpuaenabled"
      value_type    = "choice"
      value         = var.smartscreen_pua_enabled ? "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_smartscreenpuaenabled_1" : "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_smartscreenpuaenabled_0"
    }
  }

  # Site Isolation
  dynamic "setting" {
    for_each = var.site_isolation_enabled != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_sitepersiteprocess"
      value_type    = "choice"
      value         = var.site_isolation_enabled ? "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_sitepersiteprocess_1" : "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_sitepersiteprocess_0"
    }
  }

  # ============================================================================
  # Password Manager
  # ============================================================================

  # Enable Password Manager
  dynamic "setting" {
    for_each = var.password_manager_enabled != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~passwordmanager_passwordmanagerenabled"
      value_type    = "choice"
      value         = var.password_manager_enabled ? "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~passwordmanager_passwordmanagerenabled_1" : "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~passwordmanager_passwordmanagerenabled_0"
    }
  }

  # ============================================================================
  # Privacy and Tracking
  # ============================================================================

  # Tracking prevention mode
  dynamic "setting" {
    for_each = var.tracking_prevention_mode != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_trackingprevention"
      value_type    = "choice"
      value         = lookup({
        "off"      = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_trackingprevention_0"
        "basic"    = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_trackingprevention_1"
        "balanced" = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_trackingprevention_2"
        "strict"   = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_trackingprevention_3"
      }, var.tracking_prevention_mode, "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_trackingprevention_2")
    }
  }

  # Do Not Track
  dynamic "setting" {
    for_each = var.do_not_track_enabled != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_configurodonottrack"
      value_type    = "choice"
      value         = var.do_not_track_enabled ? "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_configurodonottrack_1" : "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_configurodonottrack_0"
    }
  }

  # ============================================================================
  # AI and Copilot in Edge
  # ============================================================================

  # Disable Copilot in Edge sidebar
  dynamic "setting" {
    for_each = var.disable_copilot_sidebar != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev120~policy~microsoft_edge_hubssidebarenabled"
      value_type    = "choice"
      value         = var.disable_copilot_sidebar ? "device_vendor_msft_policy_config_microsoft_edgev120~policy~microsoft_edge_hubssidebarenabled_0" : "device_vendor_msft_policy_config_microsoft_edgev120~policy~microsoft_edge_hubssidebarenabled_1"
    }
  }

  # ============================================================================
  # Extensions
  # ============================================================================

  # Control which extensions cannot be installed
  dynamic "setting" {
    for_each = var.blocked_extensions != null && length(var.blocked_extensions) > 0 ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~extensions_extensioninstallblocklist"
      value_type    = "collection"
      value         = jsonencode(var.blocked_extensions)
    }
  }

  # ============================================================================
  # Sync Settings
  # ============================================================================

  # Enable sync
  dynamic "setting" {
    for_each = var.sync_disabled != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_syncdisabled"
      value_type    = "choice"
      value         = var.sync_disabled ? "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_syncdisabled_1" : "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_syncdisabled_0"
    }
  }

  # ============================================================================
  # Pop-ups and Redirects
  # ============================================================================

  # Block pop-ups
  dynamic "setting" {
    for_each = var.block_popups != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~contentsettings_defaultpopupssetting"
      value_type    = "choice"
      value         = var.block_popups ? "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~contentsettings_defaultpopupssetting_2" : "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge~contentsettings_defaultpopupssetting_1"
    }
  }

  # ============================================================================
  # Downloads
  # ============================================================================

  # Default download directory
  dynamic "setting" {
    for_each = var.download_directory != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_microsoft_edgev80diff~policy~microsoft_edge_downloaddirectory"
      value_type    = "string"
      value         = var.download_directory
    }
  }
}
