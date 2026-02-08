# BitLocker Module Variables

variable "policy_id" {
  description = "The ID of the Settings Catalog policy to add these settings to"
  type        = string
}

variable "require_device_encryption" {
  description = "Require BitLocker encryption on the device"
  type        = bool
  default     = true
}

variable "allow_warning_for_other_disk_encryption" {
  description = "Show warning dialog for other disk encryption products"
  type        = bool
  default     = false
}

variable "allow_standard_user_encryption" {
  description = "Allow standard users to enable encryption during Azure AD Join"
  type        = bool
  default     = true
}

variable "os_drive_encryption_method" {
  description = "Encryption method for OS drive. Valid values: aes128, aes256, xts_aes128, xts_aes256"
  type        = string
  default     = "xts_aes256"

  validation {
    condition     = contains(["aes128", "aes256", "xts_aes128", "xts_aes256"], var.os_drive_encryption_method)
    error_message = "os_drive_encryption_method must be one of: aes128, aes256, xts_aes128, xts_aes256"
  }
}

variable "fixed_drive_encryption_method" {
  description = "Encryption method for fixed data drives. Valid values: aes128, aes256, xts_aes128, xts_aes256"
  type        = string
  default     = "xts_aes256"

  validation {
    condition     = contains(["aes128", "aes256", "xts_aes128", "xts_aes256"], var.fixed_drive_encryption_method)
    error_message = "fixed_drive_encryption_method must be one of: aes128, aes256, xts_aes128, xts_aes256"
  }
}

variable "removable_drive_encryption_method" {
  description = "Encryption method for removable drives. Valid values: aes128, aes256, xts_aes128, xts_aes256"
  type        = string
  default     = "aes128"

  validation {
    condition     = contains(["aes128", "aes256", "xts_aes128", "xts_aes256"], var.removable_drive_encryption_method)
    error_message = "removable_drive_encryption_method must be one of: aes128, aes256, xts_aes128, xts_aes256"
  }
}

variable "require_tpm" {
  description = "Require TPM for BitLocker startup"
  type        = bool
  default     = true
}

variable "recovery_password_rotation" {
  description = "Recovery password rotation policy. Valid values: disabled, azure_ad_only, azure_ad_and_hybrid"
  type        = string
  default     = "azure_ad_only"

  validation {
    condition     = contains(["disabled", "azure_ad_only", "azure_ad_and_hybrid"], var.recovery_password_rotation)
    error_message = "recovery_password_rotation must be one of: disabled, azure_ad_only, azure_ad_and_hybrid"
  }
}

variable "hide_recovery_options" {
  description = "Hide recovery options from the BitLocker setup wizard"
  type        = bool
  default     = true
}

variable "backup_recovery_to_azure_ad" {
  description = "Backup recovery information to Azure AD"
  type        = bool
  default     = true
}
