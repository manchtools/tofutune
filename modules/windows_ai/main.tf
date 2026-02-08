# Windows AI Settings Module
# This module provides settings for Windows AI features including Copilot, Recall, and other AI capabilities

terraform {
  required_providers {
    intune = {
      source = "MANCHTOOLS/tofutune"
    }
  }
}

resource "intune_settings_catalog_policy_settings" "windows_ai" {
  policy_id = var.policy_id

  # ============================================================================
  # Windows Copilot Settings
  # ============================================================================

  # Turn off Windows Copilot
  dynamic "setting" {
    for_each = var.disable_copilot != null ? [1] : []
    content {
      definition_id = "user_vendor_msft_policy_config_windowsai_turnoffwindowscopilot"
      value_type    = "choice"
      value         = var.disable_copilot ? "user_vendor_msft_policy_config_windowsai_turnoffwindowscopilot_1" : "user_vendor_msft_policy_config_windowsai_turnoffwindowscopilot_0"
    }
  }

  # Disable Copilot (Device level)
  dynamic "setting" {
    for_each = var.disable_copilot_device != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_windowsai_turnoffwindowscopilot"
      value_type    = "choice"
      value         = var.disable_copilot_device ? "device_vendor_msft_policy_config_windowsai_turnoffwindowscopilot_1" : "device_vendor_msft_policy_config_windowsai_turnoffwindowscopilot_0"
    }
  }

  # ============================================================================
  # Windows Recall Settings (Windows 11 24H2+)
  # ============================================================================

  # Disable Recall
  dynamic "setting" {
    for_each = var.disable_recall != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_windowsai_disableaidataanalysis"
      value_type    = "choice"
      value         = var.disable_recall ? "device_vendor_msft_policy_config_windowsai_disableaidataanalysis_1" : "device_vendor_msft_policy_config_windowsai_disableaidataanalysis_0"
    }
  }

  # Turn off saving snapshots for Windows (Recall snapshots)
  dynamic "setting" {
    for_each = var.disable_recall_snapshots != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_windowsai_turnoffsavingsnapshots"
      value_type    = "choice"
      value         = var.disable_recall_snapshots ? "device_vendor_msft_policy_config_windowsai_turnoffsavingsnapshots_1" : "device_vendor_msft_policy_config_windowsai_turnoffsavingsnapshots_0"
    }
  }

  # ============================================================================
  # Generative AI Features
  # ============================================================================

  # Disable Image Creator
  dynamic "setting" {
    for_each = var.disable_image_creator != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_windowsai_disableimagecreator"
      value_type    = "choice"
      value         = var.disable_image_creator ? "device_vendor_msft_policy_config_windowsai_disableimagecreator_1" : "device_vendor_msft_policy_config_windowsai_disableimagecreator_0"
    }
  }

  # Disable Cocreator (Paint AI features)
  dynamic "setting" {
    for_each = var.disable_cocreator != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_windowsai_disablecocreator"
      value_type    = "choice"
      value         = var.disable_cocreator ? "device_vendor_msft_policy_config_windowsai_disablecocreator_1" : "device_vendor_msft_policy_config_windowsai_disablecocreator_0"
    }
  }

  # ============================================================================
  # Microsoft 365 Copilot / Office AI Settings
  # ============================================================================

  # Allow Copilot
  dynamic "setting" {
    for_each = var.allow_copilot != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_windowsai_allowcopilot"
      value_type    = "choice"
      value         = var.allow_copilot ? "device_vendor_msft_policy_config_windowsai_allowcopilot_1" : "device_vendor_msft_policy_config_windowsai_allowcopilot_0"
    }
  }

  # ============================================================================
  # Search and Suggestions AI Features
  # ============================================================================

  # Disable AI-powered search highlights
  dynamic "setting" {
    for_each = var.disable_search_highlights != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_search_allowsearchhighlights"
      value_type    = "choice"
      value         = var.disable_search_highlights ? "device_vendor_msft_policy_config_search_allowsearchhighlights_0" : "device_vendor_msft_policy_config_search_allowsearchhighlights_1"
    }
  }

  # Disable cloud search (Bing/web results in search)
  dynamic "setting" {
    for_each = var.disable_cloud_search != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_search_allowcloudsearch"
      value_type    = "choice"
      value         = var.disable_cloud_search ? "device_vendor_msft_policy_config_search_allowcloudsearch_0" : "device_vendor_msft_policy_config_search_allowcloudsearch_1"
    }
  }

  # ============================================================================
  # Tips and Suggestions
  # ============================================================================

  # Disable tips and suggestions notifications
  dynamic "setting" {
    for_each = var.disable_tips_notifications != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_experience_allowwindowstips"
      value_type    = "choice"
      value         = var.disable_tips_notifications ? "device_vendor_msft_policy_config_experience_allowwindowstips_0" : "device_vendor_msft_policy_config_experience_allowwindowstips_1"
    }
  }

  # Disable Spotlight (lock screen suggestions)
  dynamic "setting" {
    for_each = var.disable_spotlight != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_experience_allowwindowsspotlight"
      value_type    = "choice"
      value         = var.disable_spotlight ? "device_vendor_msft_policy_config_experience_allowwindowsspotlight_0" : "device_vendor_msft_policy_config_experience_allowwindowsspotlight_1"
    }
  }

  # Disable suggested content in Settings app
  dynamic "setting" {
    for_each = var.disable_settings_suggestions != null ? [1] : []
    content {
      definition_id = "user_vendor_msft_policy_config_experience_allowwindowsspotlightonactioncenter"
      value_type    = "choice"
      value         = var.disable_settings_suggestions ? "user_vendor_msft_policy_config_experience_allowwindowsspotlightonactioncenter_0" : "user_vendor_msft_policy_config_experience_allowwindowsspotlightonactioncenter_1"
    }
  }
}
