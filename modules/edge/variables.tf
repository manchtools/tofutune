# Microsoft Edge Module Variables

variable "policy_id" {
  description = "The ID of the Settings Catalog policy to add these settings to"
  type        = string
}

# ============================================================================
# Homepage and Startup
# ============================================================================

variable "homepage_url" {
  description = "The URL to use as the homepage"
  type        = string
  default     = null
}

variable "startup_action" {
  description = "Action when Edge starts. Valid values: open_new_tab, restore_session, open_urls"
  type        = string
  default     = "open_new_tab"

  validation {
    condition     = var.startup_action == null || contains(["open_new_tab", "restore_session", "open_urls"], var.startup_action)
    error_message = "startup_action must be one of: open_new_tab, restore_session, open_urls"
  }
}

# ============================================================================
# Security Settings
# ============================================================================

variable "enable_smartscreen" {
  description = "Enable Microsoft Defender SmartScreen for Edge"
  type        = bool
  default     = true
}

variable "smartscreen_pua_enabled" {
  description = "Enable SmartScreen to block potentially unwanted apps"
  type        = bool
  default     = true
}

variable "site_isolation_enabled" {
  description = "Enable site isolation for additional security"
  type        = bool
  default     = true
}

# ============================================================================
# Password Manager
# ============================================================================

variable "password_manager_enabled" {
  description = "Enable the built-in password manager"
  type        = bool
  default     = true
}

# ============================================================================
# Privacy and Tracking
# ============================================================================

variable "tracking_prevention_mode" {
  description = "Tracking prevention level. Valid values: off, basic, balanced, strict"
  type        = string
  default     = "balanced"

  validation {
    condition     = contains(["off", "basic", "balanced", "strict"], var.tracking_prevention_mode)
    error_message = "tracking_prevention_mode must be one of: off, basic, balanced, strict"
  }
}

variable "do_not_track_enabled" {
  description = "Send Do Not Track requests"
  type        = bool
  default     = true
}

# ============================================================================
# AI and Copilot in Edge
# ============================================================================

variable "disable_copilot_sidebar" {
  description = "Disable the Copilot sidebar in Edge"
  type        = bool
  default     = null
}

# ============================================================================
# Extensions
# ============================================================================

variable "blocked_extensions" {
  description = "List of extension IDs to block. Use '*' to block all extensions."
  type        = list(string)
  default     = null
}

# ============================================================================
# Sync Settings
# ============================================================================

variable "sync_disabled" {
  description = "Disable syncing of Edge data"
  type        = bool
  default     = null
}

# ============================================================================
# Pop-ups and Redirects
# ============================================================================

variable "block_popups" {
  description = "Block pop-up windows"
  type        = bool
  default     = true
}

# ============================================================================
# Downloads
# ============================================================================

variable "download_directory" {
  description = "Default download directory path"
  type        = string
  default     = null
}
