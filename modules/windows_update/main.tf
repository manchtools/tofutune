# Windows Update for Business Settings Module
# This module provides a simple interface for Windows Update settings

terraform {
  required_providers {
    intune = {
      source = "MANCHTOOLS/tofutune"
    }
  }
}

# Create the settings for the policy
resource "intune_settings_catalog_policy_settings" "windows_update" {
  policy_id = var.policy_id

  # Automatic update behavior
  dynamic "setting" {
    for_each = var.automatic_update_mode != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_allowautoupdate"
      value_type    = "choice"
      value         = lookup({
        "notify_download"              = "device_vendor_msft_policy_config_update_allowautoupdate_2"
        "auto_download_notify_install" = "device_vendor_msft_policy_config_update_allowautoupdate_3"
        "auto_install_and_reboot_at_scheduled_time" = "device_vendor_msft_policy_config_update_allowautoupdate_4"
        "auto_install_and_reboot_without_user_control" = "device_vendor_msft_policy_config_update_allowautoupdate_5"
      }, var.automatic_update_mode, "device_vendor_msft_policy_config_update_allowautoupdate_4")
    }
  }

  # Feature update deferral (days)
  dynamic "setting" {
    for_each = var.feature_update_deferral_days != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_deferfeatureupdatesperioddays"
      value_type    = "integer"
      value         = tostring(var.feature_update_deferral_days)
    }
  }

  # Quality update deferral (days)
  dynamic "setting" {
    for_each = var.quality_update_deferral_days != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_deferqualityupdatesperioddays"
      value_type    = "integer"
      value         = tostring(var.quality_update_deferral_days)
    }
  }

  # Active hours start
  dynamic "setting" {
    for_each = var.active_hours_start != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_activehoursstart"
      value_type    = "integer"
      value         = tostring(var.active_hours_start)
    }
  }

  # Active hours end
  dynamic "setting" {
    for_each = var.active_hours_end != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_activehoursend"
      value_type    = "integer"
      value         = tostring(var.active_hours_end)
    }
  }

  # Scheduled install day (0=everyday, 1=Sunday, 7=Saturday)
  dynamic "setting" {
    for_each = var.scheduled_install_day != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_scheduledinstallday"
      value_type    = "choice"
      value         = "device_vendor_msft_policy_config_update_scheduledinstallday_${var.scheduled_install_day}"
    }
  }

  # Scheduled install time (0-23)
  dynamic "setting" {
    for_each = var.scheduled_install_time != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_scheduledinstalltime"
      value_type    = "integer"
      value         = tostring(var.scheduled_install_time)
    }
  }

  # Deadline for feature updates (days)
  dynamic "setting" {
    for_each = var.feature_update_deadline_days != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_configuredeadlineforfeatureupdates"
      value_type    = "integer"
      value         = tostring(var.feature_update_deadline_days)
    }
  }

  # Deadline for quality updates (days)
  dynamic "setting" {
    for_each = var.quality_update_deadline_days != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_configuredeadlineforqualityupdates"
      value_type    = "integer"
      value         = tostring(var.quality_update_deadline_days)
    }
  }

  # Grace period (days)
  dynamic "setting" {
    for_each = var.grace_period_days != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_configuredeadlinegraceperiod"
      value_type    = "integer"
      value         = tostring(var.grace_period_days)
    }
  }

  # Allow auto-restart outside active hours
  dynamic "setting" {
    for_each = var.auto_restart_notification_dismiss != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_autorestartrequirednotificationdismissal"
      value_type    = "choice"
      value         = var.auto_restart_notification_dismiss ? "device_vendor_msft_policy_config_update_autorestartrequirednotificationdismissal_2" : "device_vendor_msft_policy_config_update_autorestartrequirednotificationdismissal_1"
    }
  }

  # Pause feature updates
  dynamic "setting" {
    for_each = var.pause_feature_updates ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_pausefeatureupdates"
      value_type    = "choice"
      value         = "device_vendor_msft_policy_config_update_pausefeatureupdates_1"
    }
  }

  # Pause quality updates
  dynamic "setting" {
    for_each = var.pause_quality_updates ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_pausequalityupdates"
      value_type    = "choice"
      value         = "device_vendor_msft_policy_config_update_pausequalityupdates_1"
    }
  }

  # Enable preview builds
  dynamic "setting" {
    for_each = var.enable_preview_builds != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_managepreviewbuilds"
      value_type    = "choice"
      value         = lookup({
        "disabled"        = "device_vendor_msft_policy_config_update_managepreviewbuilds_0"
        "dev_channel"     = "device_vendor_msft_policy_config_update_managepreviewbuilds_2"
        "beta_channel"    = "device_vendor_msft_policy_config_update_managepreviewbuilds_3"
        "release_preview" = "device_vendor_msft_policy_config_update_managepreviewbuilds_4"
      }, var.enable_preview_builds, "device_vendor_msft_policy_config_update_managepreviewbuilds_0")
    }
  }

  # Allow Windows Update for drivers
  dynamic "setting" {
    for_each = var.allow_driver_updates != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_policy_config_update_excludewudriversinqualityupdate"
      value_type    = "choice"
      value         = var.allow_driver_updates ? "device_vendor_msft_policy_config_update_excludewudriversinqualityupdate_0" : "device_vendor_msft_policy_config_update_excludewudriversinqualityupdate_1"
    }
  }
}
