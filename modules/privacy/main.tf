# Windows Privacy Settings Module
# This module provides settings for privacy, telemetry, and data collection

terraform {
  required_providers {
    intune = {
      source = "MANCHTOOLS/tofutune"
    }
  }
}

resource "intune_settings_catalog_policy_settings" "privacy" {
  policy_id = var.policy_id

  # ============================================================================
  # Telemetry and Diagnostic Data
  # ============================================================================

  # Telemetry level (0=Security, 1=Basic, 2=Enhanced, 3=Full)
  dynamic "setting" {
    for_each = var.telemetry_level != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_system_allowtelemetry"
      value_type    = "choice"
      value         = lookup({
        "security" = "device_vendor_msft_policy_config_system_allowtelemetry_0"
        "basic"    = "device_vendor_msft_policy_config_system_allowtelemetry_1"
        "enhanced" = "device_vendor_msft_policy_config_system_allowtelemetry_2"
        "full"     = "device_vendor_msft_policy_config_system_allowtelemetry_3"
      }, var.telemetry_level, "device_vendor_msft_policy_config_system_allowtelemetry_1")
    }
  }

  # Limit diagnostic log collection
  dynamic "setting" {
    for_each = var.limit_diagnostic_log_collection != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_system_limitdiagnosticlogcollection"
      value_type    = "choice"
      value         = var.limit_diagnostic_log_collection ? "device_vendor_msft_policy_config_system_limitdiagnosticlogcollection_1" : "device_vendor_msft_policy_config_system_limitdiagnosticlogcollection_0"
    }
  }

  # Limit dump collection
  dynamic "setting" {
    for_each = var.limit_dump_collection != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_system_limitenhanceddiagnosticdatawindowsanalytics"
      value_type    = "choice"
      value         = var.limit_dump_collection ? "device_vendor_msft_policy_config_system_limitenhanceddiagnosticdatawindowsanalytics_1" : "device_vendor_msft_policy_config_system_limitenhanceddiagnosticdatawindowsanalytics_0"
    }
  }

  # ============================================================================
  # Experience and Personalization
  # ============================================================================

  # Allow tailored experiences with diagnostic data
  dynamic "setting" {
    for_each = var.allow_tailored_experiences != null ? [1] : []
    content {
      definition_id = "user_vendor_msft_policy_config_experience_allowtailoredexperienceswithdiagnosticdata"
      value_type    = "choice"
      value         = var.allow_tailored_experiences ? "user_vendor_msft_policy_config_experience_allowtailoredexperienceswithdiagnosticdata_1" : "user_vendor_msft_policy_config_experience_allowtailoredexperienceswithdiagnosticdata_0"
    }
  }

  # Disable advertising ID
  dynamic "setting" {
    for_each = var.disable_advertising_id != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_privacy_disableadvertisingid"
      value_type    = "choice"
      value         = var.disable_advertising_id ? "device_vendor_msft_policy_config_privacy_disableadvertisingid_1" : "device_vendor_msft_policy_config_privacy_disableadvertisingid_0"
    }
  }

  # ============================================================================
  # Location Services
  # ============================================================================

  # Allow location
  dynamic "setting" {
    for_each = var.allow_location != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_privacy_letappacccesslocation"
      value_type    = "choice"
      value         = lookup({
        "deny"         = "device_vendor_msft_policy_config_privacy_letappacccesslocation_0"
        "user_control" = "device_vendor_msft_policy_config_privacy_letappacccesslocation_1"
        "allow"        = "device_vendor_msft_policy_config_privacy_letappacccesslocation_2"
      }, var.allow_location, "device_vendor_msft_policy_config_privacy_letappacccesslocation_1")
    }
  }

  # ============================================================================
  # Camera and Microphone
  # ============================================================================

  # Allow camera
  dynamic "setting" {
    for_each = var.allow_camera != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_privacy_letappsaccesscamera"
      value_type    = "choice"
      value         = lookup({
        "deny"         = "device_vendor_msft_policy_config_privacy_letappsaccesscamera_0"
        "user_control" = "device_vendor_msft_policy_config_privacy_letappsaccesscamera_1"
        "allow"        = "device_vendor_msft_policy_config_privacy_letappsaccesscamera_2"
      }, var.allow_camera, "device_vendor_msft_policy_config_privacy_letappsaccesscamera_1")
    }
  }

  # Allow microphone
  dynamic "setting" {
    for_each = var.allow_microphone != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_privacy_letappsaccessmicrophone"
      value_type    = "choice"
      value         = lookup({
        "deny"         = "device_vendor_msft_policy_config_privacy_letappsaccessmicrophone_0"
        "user_control" = "device_vendor_msft_policy_config_privacy_letappsaccessmicrophone_1"
        "allow"        = "device_vendor_msft_policy_config_privacy_letappsaccessmicrophone_2"
      }, var.allow_microphone, "device_vendor_msft_policy_config_privacy_letappsaccessmicrophone_1")
    }
  }

  # ============================================================================
  # Activity History
  # ============================================================================

  # Enable activity feed
  dynamic "setting" {
    for_each = var.enable_activity_feed != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_privacy_enableactivityfeed"
      value_type    = "choice"
      value         = var.enable_activity_feed ? "device_vendor_msft_policy_config_privacy_enableactivityfeed_1" : "device_vendor_msft_policy_config_privacy_enableactivityfeed_0"
    }
  }

  # Publish user activities
  dynamic "setting" {
    for_each = var.publish_user_activities != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_privacy_publishuseractivities"
      value_type    = "choice"
      value         = var.publish_user_activities ? "device_vendor_msft_policy_config_privacy_publishuseractivities_1" : "device_vendor_msft_policy_config_privacy_publishuseractivities_0"
    }
  }

  # Upload user activities
  dynamic "setting" {
    for_each = var.upload_user_activities != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_privacy_uploaduseractivities"
      value_type    = "choice"
      value         = var.upload_user_activities ? "device_vendor_msft_policy_config_privacy_uploaduseractivities_1" : "device_vendor_msft_policy_config_privacy_uploaduseractivities_0"
    }
  }

  # ============================================================================
  # Customer Experience Improvement Program (CEIP)
  # ============================================================================

  # Allow device name to be sent in Windows diagnostic data
  dynamic "setting" {
    for_each = var.allow_device_name_in_telemetry != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_system_allowdevicenameindiagnosticdata"
      value_type    = "choice"
      value         = var.allow_device_name_in_telemetry ? "device_vendor_msft_policy_config_system_allowdevicenameindiagnosticdata_1" : "device_vendor_msft_policy_config_system_allowdevicenameindiagnosticdata_0"
    }
  }

  # ============================================================================
  # Feedback
  # ============================================================================

  # Disable feedback notifications
  dynamic "setting" {
    for_each = var.disable_feedback_notifications != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_experience_donotshowfeedbacknotifications"
      value_type    = "choice"
      value         = var.disable_feedback_notifications ? "device_vendor_msft_policy_config_experience_donotshowfeedbacknotifications_1" : "device_vendor_msft_policy_config_experience_donotshowfeedbacknotifications_0"
    }
  }
}
