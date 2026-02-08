# Microsoft Defender Settings Module
# This module provides a simple interface for common Microsoft Defender Antivirus settings

terraform {
  required_providers {
    intune = {
      source = "tofutune/intune"
    }
  }
}

# Create the settings for the policy
resource "intune_settings_catalog_policy_settings" "defender" {
  policy_id = var.policy_id

  # Real-time protection
  dynamic "setting" {
    for_each = var.real_time_protection != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_disablerealtimemonitoring"
      value_type    = "boolean"
      value         = var.real_time_protection ? "false" : "true"  # Note: The setting is "disable", so we invert
    }
  }

  # Behavior monitoring
  dynamic "setting" {
    for_each = var.behavior_monitoring != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_allowbehaviormonitoring"
      value_type    = "choice"
      value         = var.behavior_monitoring ? "device_vendor_msft_defender_configuration_allowbehaviormonitoring_1" : "device_vendor_msft_defender_configuration_allowbehaviormonitoring_0"
    }
  }

  # Cloud protection (MAPS)
  dynamic "setting" {
    for_each = var.cloud_protection != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_allowcloudprotection"
      value_type    = "choice"
      value         = var.cloud_protection == "enabled" ? "device_vendor_msft_defender_configuration_allowcloudprotection_1" : var.cloud_protection == "advanced" ? "device_vendor_msft_defender_configuration_allowcloudprotection_2" : "device_vendor_msft_defender_configuration_allowcloudprotection_0"
    }
  }

  # Cloud block level
  dynamic "setting" {
    for_each = var.cloud_block_level != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_cloudblocklevel"
      value_type    = "choice"
      value         = lookup({
        "default"     = "device_vendor_msft_defender_configuration_cloudblocklevel_0"
        "moderate"    = "device_vendor_msft_defender_configuration_cloudblocklevel_2"
        "high"        = "device_vendor_msft_defender_configuration_cloudblocklevel_4"
        "high_plus"   = "device_vendor_msft_defender_configuration_cloudblocklevel_6"
        "zero_tolerance" = "device_vendor_msft_defender_configuration_cloudblocklevel_99"
      }, var.cloud_block_level, "device_vendor_msft_defender_configuration_cloudblocklevel_0")
    }
  }

  # PUA (Potentially Unwanted Application) protection
  dynamic "setting" {
    for_each = var.pua_protection != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_puaprotection"
      value_type    = "choice"
      value         = lookup({
        "disabled" = "device_vendor_msft_defender_configuration_puaprotection_0"
        "enabled"  = "device_vendor_msft_defender_configuration_puaprotection_1"
        "audit"    = "device_vendor_msft_defender_configuration_puaprotection_2"
      }, var.pua_protection, "device_vendor_msft_defender_configuration_puaprotection_0")
    }
  }

  # Script scanning
  dynamic "setting" {
    for_each = var.script_scanning != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_allowscriptscanning"
      value_type    = "choice"
      value         = var.script_scanning ? "device_vendor_msft_defender_configuration_allowscriptscanning_1" : "device_vendor_msft_defender_configuration_allowscriptscanning_0"
    }
  }

  # Archive scanning
  dynamic "setting" {
    for_each = var.archive_scanning != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_allowarchivescanning"
      value_type    = "choice"
      value         = var.archive_scanning ? "device_vendor_msft_defender_configuration_allowarchivescanning_1" : "device_vendor_msft_defender_configuration_allowarchivescanning_0"
    }
  }

  # Network protection
  dynamic "setting" {
    for_each = var.network_protection != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_enablenetworkprotection"
      value_type    = "choice"
      value         = lookup({
        "disabled" = "device_vendor_msft_defender_configuration_enablenetworkprotection_0"
        "enabled"  = "device_vendor_msft_defender_configuration_enablenetworkprotection_1"
        "audit"    = "device_vendor_msft_defender_configuration_enablenetworkprotection_2"
      }, var.network_protection, "device_vendor_msft_defender_configuration_enablenetworkprotection_0")
    }
  }

  # Intrusion Prevention System (Network Inspection)
  dynamic "setting" {
    for_each = var.intrusion_prevention != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_allowintrusionpreventionsystem"
      value_type    = "choice"
      value         = var.intrusion_prevention ? "device_vendor_msft_defender_configuration_allowintrusionpreventionsystem_1" : "device_vendor_msft_defender_configuration_allowintrusionpreventionsystem_0"
    }
  }

  # Scan parameters - Schedule day
  dynamic "setting" {
    for_each = var.scheduled_scan_day != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_scheduledscanday"
      value_type    = "choice"
      value         = "device_vendor_msft_defender_configuration_scheduledscanday_${var.scheduled_scan_day}"
    }
  }

  # Scan parameters - Schedule time
  dynamic "setting" {
    for_each = var.scheduled_scan_time != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_scheduledscantime"
      value_type    = "integer"
      value         = tostring(var.scheduled_scan_time)
    }
  }

  # Scan parameters - Scan type (quick or full)
  dynamic "setting" {
    for_each = var.scheduled_scan_type != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_defender_configuration_scheduledscantype"
      value_type    = "choice"
      value         = var.scheduled_scan_type == "quick" ? "device_vendor_msft_defender_configuration_scheduledscantype_1" : "device_vendor_msft_defender_configuration_scheduledscantype_2"
    }
  }
}
