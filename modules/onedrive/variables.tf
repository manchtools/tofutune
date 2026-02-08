# OneDrive Module Variables

variable "policy_id" {
  description = "The ID of the Settings Catalog policy to add these settings to"
  type        = string
}

variable "tenant_id" {
  description = "Your Azure AD tenant ID (required for Known Folder Move)"
  type        = string
  default     = null
}

# ============================================================================
# Known Folder Move (KFM) / Folder Backup
# ============================================================================

variable "silently_move_known_folders" {
  description = "Silently move Documents, Desktop, and Pictures folders to OneDrive without user interaction"
  type        = bool
  default     = false
}

variable "prompt_to_move_known_folders" {
  description = "Prompt users to move their Documents, Desktop, and Pictures folders to OneDrive"
  type        = bool
  default     = false
}

variable "prevent_redirect_to_pc" {
  description = "Prevent users from redirecting known folders back to their PC after KFM"
  type        = bool
  default     = true
}

# ============================================================================
# Sync Settings
# ============================================================================

variable "silent_account_config" {
  description = "Silently sign in users to OneDrive sync with their Windows credentials"
  type        = bool
  default     = true
}

variable "files_on_demand_enabled" {
  description = "Enable Files On-Demand (files stored in cloud, downloaded when accessed)"
  type        = bool
  default     = true
}

# ============================================================================
# Storage and Limits
# ============================================================================

variable "max_size_before_download_gb" {
  description = "Maximum OneDrive size (GB) that will automatically download. Larger OneDrives use Files On-Demand."
  type        = number
  default     = null

  validation {
    condition     = var.max_size_before_download_gb == null || var.max_size_before_download_gb > 0
    error_message = "max_size_before_download_gb must be greater than 0"
  }
}

# ============================================================================
# Network Settings
# ============================================================================

variable "download_bandwidth_limit_percent" {
  description = "Limit download bandwidth as percentage of available bandwidth (1-99)"
  type        = number
  default     = null

  validation {
    condition     = var.download_bandwidth_limit_percent == null || (var.download_bandwidth_limit_percent >= 1 && var.download_bandwidth_limit_percent <= 99)
    error_message = "download_bandwidth_limit_percent must be between 1 and 99"
  }
}

variable "upload_bandwidth_limit_kbps" {
  description = "Limit upload bandwidth in KB/s (50-100000)"
  type        = number
  default     = null

  validation {
    condition     = var.upload_bandwidth_limit_kbps == null || (var.upload_bandwidth_limit_kbps >= 50 && var.upload_bandwidth_limit_kbps <= 100000)
    error_message = "upload_bandwidth_limit_kbps must be between 50 and 100000"
  }
}

# ============================================================================
# Sharing and Collaboration
# ============================================================================

variable "block_personal_account_sync" {
  description = "Block users from syncing personal OneDrive accounts (only allow work/school accounts)"
  type        = bool
  default     = false
}

variable "allowed_tenant_ids" {
  description = "List of Azure AD tenant IDs that are allowed to sync. Leave empty to allow all."
  type        = list(string)
  default     = null
}

variable "blocked_tenant_ids" {
  description = "List of Azure AD tenant IDs that are blocked from syncing"
  type        = list(string)
  default     = null
}

# ============================================================================
# Update Settings
# ============================================================================

variable "update_ring" {
  description = "OneDrive update ring. Valid values: enterprise, production, insiders, deferred"
  type        = string
  default     = "production"

  validation {
    condition     = contains(["enterprise", "production", "insiders", "deferred"], var.update_ring)
    error_message = "update_ring must be one of: enterprise, production, insiders, deferred"
  }
}

# ============================================================================
# User Experience
# ============================================================================

variable "prevent_network_traffic_until_signin" {
  description = "Prevent OneDrive from generating network traffic until user signs in"
  type        = bool
  default     = false
}

variable "hide_desktop_shortcut" {
  description = "Hide the OneDrive shortcut on the desktop"
  type        = bool
  default     = null
}
