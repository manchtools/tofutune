# Windows Privacy Module Variables

variable "policy_id" {
  description = "The ID of the Settings Catalog policy to add these settings to"
  type        = string
}

# ============================================================================
# Telemetry and Diagnostic Data
# ============================================================================

variable "telemetry_level" {
  description = "Windows telemetry/diagnostic data level. Valid values: security (Enterprise only), basic, enhanced, full"
  type        = string
  default     = "basic"

  validation {
    condition     = contains(["security", "basic", "enhanced", "full"], var.telemetry_level)
    error_message = "telemetry_level must be one of: security, basic, enhanced, full"
  }
}

variable "limit_diagnostic_log_collection" {
  description = "Limit the collection of diagnostic logs"
  type        = bool
  default     = true
}

variable "limit_dump_collection" {
  description = "Limit the collection of memory dumps for Windows Analytics"
  type        = bool
  default     = true
}

variable "allow_device_name_in_telemetry" {
  description = "Allow device name to be included in diagnostic data sent to Microsoft"
  type        = bool
  default     = false
}

# ============================================================================
# Experience and Personalization
# ============================================================================

variable "allow_tailored_experiences" {
  description = "Allow Microsoft to use diagnostic data to provide personalized tips and recommendations"
  type        = bool
  default     = false
}

variable "disable_advertising_id" {
  description = "Disable the advertising ID for personalized ads"
  type        = bool
  default     = true
}

# ============================================================================
# Location Services
# ============================================================================

variable "allow_location" {
  description = "Control app access to device location. Valid values: deny, user_control, allow"
  type        = string
  default     = "user_control"

  validation {
    condition     = contains(["deny", "user_control", "allow"], var.allow_location)
    error_message = "allow_location must be one of: deny, user_control, allow"
  }
}

# ============================================================================
# Camera and Microphone
# ============================================================================

variable "allow_camera" {
  description = "Control app access to camera. Valid values: deny, user_control, allow"
  type        = string
  default     = "user_control"

  validation {
    condition     = contains(["deny", "user_control", "allow"], var.allow_camera)
    error_message = "allow_camera must be one of: deny, user_control, allow"
  }
}

variable "allow_microphone" {
  description = "Control app access to microphone. Valid values: deny, user_control, allow"
  type        = string
  default     = "user_control"

  validation {
    condition     = contains(["deny", "user_control", "allow"], var.allow_microphone)
    error_message = "allow_microphone must be one of: deny, user_control, allow"
  }
}

# ============================================================================
# Activity History
# ============================================================================

variable "enable_activity_feed" {
  description = "Enable Activity Feed (timeline feature)"
  type        = bool
  default     = false
}

variable "publish_user_activities" {
  description = "Allow publishing of user activities to the activity feed"
  type        = bool
  default     = false
}

variable "upload_user_activities" {
  description = "Allow uploading user activities to Microsoft cloud"
  type        = bool
  default     = false
}

# ============================================================================
# Feedback
# ============================================================================

variable "disable_feedback_notifications" {
  description = "Disable Windows feedback notifications asking for user feedback"
  type        = bool
  default     = true
}
