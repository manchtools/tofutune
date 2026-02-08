# BitLocker Encryption Settings Module
# This module provides a simple interface for common BitLocker settings

terraform {
  required_providers {
    intune = {
      source = "tofutune/intune"
    }
  }
}

# Create the settings for the policy
resource "intune_settings_catalog_policy_settings" "bitlocker" {
  policy_id = var.policy_id

  # Require device encryption
  dynamic "setting" {
    for_each = var.require_device_encryption ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_requiredeviceencryption"
      value_type    = "choice"
      value         = "device_vendor_msft_bitlocker_requiredeviceencryption_1"
    }
  }

  # Allow warning for other disk encryption
  dynamic "setting" {
    for_each = var.allow_warning_for_other_disk_encryption != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_allowwarningforotherdiskencryption"
      value_type    = "choice"
      value         = var.allow_warning_for_other_disk_encryption ? "device_vendor_msft_bitlocker_allowwarningforotherdiskencryption_1" : "device_vendor_msft_bitlocker_allowwarningforotherdiskencryption_0"
    }
  }

  # Allow standard user encryption
  dynamic "setting" {
    for_each = var.allow_standard_user_encryption != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_allowstandarduserencryption"
      value_type    = "choice"
      value         = var.allow_standard_user_encryption ? "device_vendor_msft_bitlocker_allowstandarduserencryption_1" : "device_vendor_msft_bitlocker_allowstandarduserencryption_0"
    }
  }

  # OS Drive encryption method
  dynamic "setting" {
    for_each = var.os_drive_encryption_method != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_osdrivesencryptiontype"
      value_type    = "choice"
      value         = lookup({
        "aes128"     = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_osdrivesencryptiontype_3"
        "aes256"     = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_osdrivesencryptiontype_4"
        "xts_aes128" = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_osdrivesencryptiontype_6"
        "xts_aes256" = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_osdrivesencryptiontype_7"
      }, var.os_drive_encryption_method, "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_osdrivesencryptiontype_7")
    }
  }

  # Fixed drive encryption method
  dynamic "setting" {
    for_each = var.fixed_drive_encryption_method != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_fixeddatavolumesencryptiontype"
      value_type    = "choice"
      value         = lookup({
        "aes128"     = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_fixeddatavolumesencryptiontype_3"
        "aes256"     = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_fixeddatavolumesencryptiontype_4"
        "xts_aes128" = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_fixeddatavolumesencryptiontype_6"
        "xts_aes256" = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_fixeddatavolumesencryptiontype_7"
      }, var.fixed_drive_encryption_method, "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_fixeddatavolumesencryptiontype_7")
    }
  }

  # Removable drive encryption method
  dynamic "setting" {
    for_each = var.removable_drive_encryption_method != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_removabledatavolumesencryptiontype"
      value_type    = "choice"
      value         = lookup({
        "aes128"     = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_removabledatavolumesencryptiontype_3"
        "aes256"     = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_removabledatavolumesencryptiontype_4"
        "xts_aes128" = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_removabledatavolumesencryptiontype_6"
        "xts_aes256" = "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_removabledatavolumesencryptiontype_7"
      }, var.removable_drive_encryption_method, "device_vendor_msft_bitlocker_encryptionmethodbydrivetype_removabledatavolumesencryptiontype_3")
    }
  }

  # Startup authentication - Require TPM
  dynamic "setting" {
    for_each = var.require_tpm ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_systemdrivesrequirestartupauthentication_systemdrivesminimumpinenabled"
      value_type    = "choice"
      value         = "device_vendor_msft_bitlocker_systemdrivesrequirestartupauthentication_systemdrivesminimumpinenabled_0"
    }
  }

  # Configure recovery password rotation
  dynamic "setting" {
    for_each = var.recovery_password_rotation != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_configurerecoverypasswordrotation"
      value_type    = "choice"
      value         = lookup({
        "disabled"             = "device_vendor_msft_bitlocker_configurerecoverypasswordrotation_0"
        "azure_ad_only"        = "device_vendor_msft_bitlocker_configurerecoverypasswordrotation_1"
        "azure_ad_and_hybrid"  = "device_vendor_msft_bitlocker_configurerecoverypasswordrotation_2"
      }, var.recovery_password_rotation, "device_vendor_msft_bitlocker_configurerecoverypasswordrotation_0")
    }
  }

  # Hide recovery options from BitLocker setup wizard
  dynamic "setting" {
    for_each = var.hide_recovery_options != null ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_systemdrivesrecoveryoptions_systemdriveshiderecoveryoptions"
      value_type    = "choice"
      value         = var.hide_recovery_options ? "device_vendor_msft_bitlocker_systemdrivesrecoveryoptions_systemdriveshiderecoveryoptions_true" : "device_vendor_msft_bitlocker_systemdrivesrecoveryoptions_systemdriveshiderecoveryoptions_false"
    }
  }

  # Save BitLocker recovery information to Azure AD
  dynamic "setting" {
    for_each = var.backup_recovery_to_azure_ad ? [1] : []
    content {
      definition_id = "device_vendor_msft_bitlocker_systemdrivesrecoveryoptions_systemdrivesactivebackup"
      value_type    = "choice"
      value         = "device_vendor_msft_bitlocker_systemdrivesrecoveryoptions_systemdrivesactivebackup_true"
    }
  }
}
