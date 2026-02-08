# Windows Update Module Variables

variable "policy_id" {
  description = "The ID of the Settings Catalog policy to add these settings to"
  type        = string
}

variable "automatic_update_mode" {
  description = "Automatic update behavior. Valid values: notify_download, auto_download_notify_install, auto_install_and_reboot_at_scheduled_time, auto_install_and_reboot_without_user_control"
  type        = string
  default     = "auto_install_and_reboot_at_scheduled_time"

  validation {
    condition     = contains(["notify_download", "auto_download_notify_install", "auto_install_and_reboot_at_scheduled_time", "auto_install_and_reboot_without_user_control"], var.automatic_update_mode)
    error_message = "automatic_update_mode must be one of the valid values"
  }
}

variable "feature_update_deferral_days" {
  description = "Number of days to defer feature updates (0-365)"
  type        = number
  default     = 0

  validation {
    condition     = var.feature_update_deferral_days >= 0 && var.feature_update_deferral_days <= 365
    error_message = "feature_update_deferral_days must be between 0 and 365"
  }
}

variable "quality_update_deferral_days" {
  description = "Number of days to defer quality updates (0-30)"
  type        = number
  default     = 0

  validation {
    condition     = var.quality_update_deferral_days >= 0 && var.quality_update_deferral_days <= 30
    error_message = "quality_update_deferral_days must be between 0 and 30"
  }
}

variable "active_hours_start" {
  description = "Start of active hours (0-23)"
  type        = number
  default     = 8

  validation {
    condition     = var.active_hours_start >= 0 && var.active_hours_start <= 23
    error_message = "active_hours_start must be between 0 and 23"
  }
}

variable "active_hours_end" {
  description = "End of active hours (0-23)"
  type        = number
  default     = 17

  validation {
    condition     = var.active_hours_end >= 0 && var.active_hours_end <= 23
    error_message = "active_hours_end must be between 0 and 23"
  }
}

variable "scheduled_install_day" {
  description = "Day for scheduled installation. 0=everyday, 1=Sunday, 2=Monday, ..., 7=Saturday"
  type        = number
  default     = null

  validation {
    condition     = var.scheduled_install_day == null || (var.scheduled_install_day >= 0 && var.scheduled_install_day <= 7)
    error_message = "scheduled_install_day must be between 0 and 7"
  }
}

variable "scheduled_install_time" {
  description = "Hour for scheduled installation (0-23)"
  type        = number
  default     = null

  validation {
    condition     = var.scheduled_install_time == null || (var.scheduled_install_time >= 0 && var.scheduled_install_time <= 23)
    error_message = "scheduled_install_time must be between 0 and 23"
  }
}

variable "feature_update_deadline_days" {
  description = "Deadline for feature updates in days (2-30)"
  type        = number
  default     = 7

  validation {
    condition     = var.feature_update_deadline_days >= 2 && var.feature_update_deadline_days <= 30
    error_message = "feature_update_deadline_days must be between 2 and 30"
  }
}

variable "quality_update_deadline_days" {
  description = "Deadline for quality updates in days (2-30)"
  type        = number
  default     = 3

  validation {
    condition     = var.quality_update_deadline_days >= 2 && var.quality_update_deadline_days <= 30
    error_message = "quality_update_deadline_days must be between 2 and 30"
  }
}

variable "grace_period_days" {
  description = "Grace period before forced restart in days (0-7)"
  type        = number
  default     = 2

  validation {
    condition     = var.grace_period_days >= 0 && var.grace_period_days <= 7
    error_message = "grace_period_days must be between 0 and 7"
  }
}

variable "auto_restart_notification_dismiss" {
  description = "Allow auto-dismiss of restart notification"
  type        = bool
  default     = false
}

variable "pause_feature_updates" {
  description = "Pause feature updates"
  type        = bool
  default     = false
}

variable "pause_quality_updates" {
  description = "Pause quality updates"
  type        = bool
  default     = false
}

variable "enable_preview_builds" {
  description = "Enable preview builds. Valid values: disabled, dev_channel, beta_channel, release_preview"
  type        = string
  default     = "disabled"

  validation {
    condition     = contains(["disabled", "dev_channel", "beta_channel", "release_preview"], var.enable_preview_builds)
    error_message = "enable_preview_builds must be one of: disabled, dev_channel, beta_channel, release_preview"
  }
}

variable "allow_driver_updates" {
  description = "Allow driver updates from Windows Update"
  type        = bool
  default     = true
}
