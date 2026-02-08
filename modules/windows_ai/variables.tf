# Windows AI Module Variables

variable "policy_id" {
  description = "The ID of the Settings Catalog policy to add these settings to"
  type        = string
}

# ============================================================================
# Windows Copilot
# ============================================================================

variable "disable_copilot" {
  description = "Disable Windows Copilot for users. When true, Copilot icon is removed from taskbar and users cannot access it."
  type        = bool
  default     = null
}

variable "disable_copilot_device" {
  description = "Disable Windows Copilot at device level (applies to all users on the device)"
  type        = bool
  default     = null
}

variable "allow_copilot" {
  description = "Explicitly allow or block Copilot. Use false to block Copilot entirely."
  type        = bool
  default     = null
}

# ============================================================================
# Windows Recall (Windows 11 24H2+)
# ============================================================================

variable "disable_recall" {
  description = "Disable Windows Recall AI data analysis feature. When true, prevents Recall from analyzing and storing activity data."
  type        = bool
  default     = null
}

variable "disable_recall_snapshots" {
  description = "Disable saving snapshots for Recall. When true, Recall will not capture screenshots of user activity."
  type        = bool
  default     = null
}

# ============================================================================
# Generative AI Features
# ============================================================================

variable "disable_image_creator" {
  description = "Disable Image Creator (AI image generation). When true, Image Creator features are disabled in Paint and other apps."
  type        = bool
  default     = null
}

variable "disable_cocreator" {
  description = "Disable Cocreator (AI features in Paint). When true, AI-powered drawing assistance is disabled."
  type        = bool
  default     = null
}

# ============================================================================
# Search and Suggestions
# ============================================================================

variable "disable_search_highlights" {
  description = "Disable AI-powered search highlights. When true, removes trending searches and highlights from Windows Search."
  type        = bool
  default     = null
}

variable "disable_cloud_search" {
  description = "Disable cloud/web search results. When true, Windows Search only shows local results (no Bing integration)."
  type        = bool
  default     = null
}

# ============================================================================
# Tips and Spotlight
# ============================================================================

variable "disable_tips_notifications" {
  description = "Disable Windows Tips and suggestions notifications"
  type        = bool
  default     = null
}

variable "disable_spotlight" {
  description = "Disable Windows Spotlight (rotating images and suggestions on lock screen)"
  type        = bool
  default     = null
}

variable "disable_settings_suggestions" {
  description = "Disable suggested content and tips in the Settings app"
  type        = bool
  default     = null
}
