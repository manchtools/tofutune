# Microsoft Defender Module Variables

variable "policy_id" {
  description = "The ID of the Settings Catalog policy to add these settings to"
  type        = string
}

variable "real_time_protection" {
  description = "Enable real-time protection. When enabled, Defender actively monitors for malware in real-time."
  type        = bool
  default     = true
}

variable "behavior_monitoring" {
  description = "Enable behavior monitoring. Monitors processes for suspicious behavior."
  type        = bool
  default     = true
}

variable "cloud_protection" {
  description = "Cloud protection level. Valid values: disabled, enabled, advanced"
  type        = string
  default     = "enabled"

  validation {
    condition     = contains(["disabled", "enabled", "advanced"], var.cloud_protection)
    error_message = "cloud_protection must be one of: disabled, enabled, advanced"
  }
}

variable "cloud_block_level" {
  description = "Cloud blocking level for unknown files. Valid values: default, moderate, high, high_plus, zero_tolerance"
  type        = string
  default     = "high"

  validation {
    condition     = contains(["default", "moderate", "high", "high_plus", "zero_tolerance"], var.cloud_block_level)
    error_message = "cloud_block_level must be one of: default, moderate, high, high_plus, zero_tolerance"
  }
}

variable "pua_protection" {
  description = "Potentially Unwanted Application (PUA) protection. Valid values: disabled, enabled, audit"
  type        = string
  default     = "enabled"

  validation {
    condition     = contains(["disabled", "enabled", "audit"], var.pua_protection)
    error_message = "pua_protection must be one of: disabled, enabled, audit"
  }
}

variable "script_scanning" {
  description = "Enable scanning of scripts (PowerShell, VBS, etc.)"
  type        = bool
  default     = true
}

variable "archive_scanning" {
  description = "Enable scanning inside archive files (ZIP, CAB, etc.)"
  type        = bool
  default     = true
}

variable "network_protection" {
  description = "Network protection mode. Valid values: disabled, enabled, audit"
  type        = string
  default     = "enabled"

  validation {
    condition     = contains(["disabled", "enabled", "audit"], var.network_protection)
    error_message = "network_protection must be one of: disabled, enabled, audit"
  }
}

variable "intrusion_prevention" {
  description = "Enable network intrusion prevention system"
  type        = bool
  default     = true
}

variable "scheduled_scan_day" {
  description = "Day of week for scheduled scan. 0=Everyday, 1=Sunday, 2=Monday, ..., 7=Saturday, 8=No scheduled scan"
  type        = number
  default     = null

  validation {
    condition     = var.scheduled_scan_day == null || (var.scheduled_scan_day >= 0 && var.scheduled_scan_day <= 8)
    error_message = "scheduled_scan_day must be between 0 and 8"
  }
}

variable "scheduled_scan_time" {
  description = "Time for scheduled scan in minutes from midnight (e.g., 120 = 2:00 AM)"
  type        = number
  default     = null

  validation {
    condition     = var.scheduled_scan_time == null || (var.scheduled_scan_time >= 0 && var.scheduled_scan_time <= 1439)
    error_message = "scheduled_scan_time must be between 0 and 1439 (minutes from midnight)"
  }
}

variable "scheduled_scan_type" {
  description = "Type of scheduled scan. Valid values: quick, full"
  type        = string
  default     = null

  validation {
    condition     = var.scheduled_scan_type == null || contains(["quick", "full"], var.scheduled_scan_type)
    error_message = "scheduled_scan_type must be one of: quick, full"
  }
}
