# TofuTune - OpenTofu Provider for Microsoft Intune

TofuTune is an OpenTofu/Terraform provider for managing Microsoft Intune device management configurations. It focuses on configuration management (not applications) and provides a modular approach to building Intune policies.

## Features

- **Azure AD-Compatible Authentication**: Uses the same authentication methods as the `hashicorp/azuread` provider
- **Settings Catalog Policies**: Full support for the Intune Settings Catalog with thousands of Windows settings
- **Modular Architecture**: Build policies from reusable modules for Defender, BitLocker, Windows Update, and more
- **Compliance Policies**: Define and enforce device compliance requirements
- **Endpoint Security**: Configure endpoint security policies for antivirus, firewall, disk encryption
- **Policy Assignments**: Assign policies to Azure AD groups with include/exclude support

## Quick Start

### Installation

```hcl
terraform {
  required_providers {
    intune = {
      source  = "MANCHTOOLS/tofutune"
      version = "~> 0.1"
    }
  }
}

provider "intune" {
  # Uses ARM_TENANT_ID, ARM_CLIENT_ID, ARM_CLIENT_SECRET from environment
  # Or use Azure CLI: run 'az login' first and set use_cli = true
}
```

### Basic Usage

```hcl
# Create a Settings Catalog policy
resource "intune_settings_catalog_policy" "security" {
  name         = "Windows Security Baseline"
  platforms    = "windows10AndLater"
  technologies = "mdm"
}

# Add settings using the Defender module
module "defender" {
  source    = "MANCHTOOLS/tofutune//modules/defender"
  policy_id = intune_settings_catalog_policy.security.id

  real_time_protection = true
  cloud_protection     = "enabled"
  network_protection   = "enabled"
}

# Assign to all devices
resource "intune_policy_assignment" "security" {
  policy_id   = intune_settings_catalog_policy.security.id
  policy_type = "settings_catalog"
  all_devices = true
}
```

## Authentication

The provider supports the same authentication methods as `hashicorp/azuread`:

### Azure CLI

```hcl
provider "intune" {
  use_cli = true
}
```

Run `az login` before using the provider.

### Service Principal with Client Secret

```hcl
provider "intune" {
  tenant_id     = "00000000-0000-0000-0000-000000000000"
  client_id     = "00000000-0000-0000-0000-000000000000"
  client_secret = "your-client-secret"
}
```

Or use environment variables:
- `ARM_TENANT_ID`
- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`

### Service Principal with Client Certificate

```hcl
provider "intune" {
  tenant_id              = "00000000-0000-0000-0000-000000000000"
  client_id              = "00000000-0000-0000-0000-000000000000"
  client_certificate_path = "/path/to/certificate.pem"
}
```

### Managed Identity

```hcl
provider "intune" {
  use_msi = true
  # Optionally specify client ID for user-assigned managed identity
  # msi_client_id = "00000000-0000-0000-0000-000000000000"
}
```

### OIDC (GitHub Actions)

```hcl
provider "intune" {
  use_oidc   = true
  tenant_id  = "00000000-0000-0000-0000-000000000000"
  client_id  = "00000000-0000-0000-0000-000000000000"
}
```

## Required Permissions

The service principal or user must have the following Microsoft Graph API permissions:

| Permission | Type | Description |
|------------|------|-------------|
| `DeviceManagementConfiguration.ReadWrite.All` | Application | Settings Catalog policies, assignment filters |
| `DeviceManagementManagedDevices.ReadWrite.All` | Application | Compliance policies |
| `DeviceManagementRBAC.ReadWrite.All` | Application | Scope tags |
| `Group.Read.All` | Application | Read groups for assignments |

## Resources

### Core Resources

| Resource | Description |
|----------|-------------|
| `intune_settings_catalog_policy` | Settings Catalog policy container |
| `intune_settings_catalog_policy_settings` | Settings within a policy (modular) |
| `intune_compliance_policy` | Device compliance policy (Windows 10/11) |
| `intune_endpoint_security_policy` | Endpoint security policy |
| `intune_policy_assignment` | Policy assignment to groups |
| `intune_scope_tag` | Role scope tag for RBAC |
| `intune_assignment_filter` | Assignment filter for dynamic device targeting |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| `intune_setting_definition` | Look up setting definition IDs |
| `intune_settings_catalog_template` | Look up template information |
| `intune_policy` | Read existing policies |
| `intune_scope_tags` | List all scope tags |
| `intune_assignment_filters` | List assignment filters |

## Pre-built Modules

The provider includes pre-built modules for common configurations:

### Microsoft Defender

```hcl
module "defender" {
  source    = "MANCHTOOLS/tofutune//modules/defender"
  policy_id = intune_settings_catalog_policy.example.id

  real_time_protection = true
  behavior_monitoring  = true
  cloud_protection     = "enabled"     # disabled, enabled, advanced
  cloud_block_level    = "high"        # default, moderate, high, high_plus, zero_tolerance
  pua_protection       = "enabled"     # disabled, enabled, audit
  network_protection   = "enabled"     # disabled, enabled, audit
  intrusion_prevention = true
  scheduled_scan_day   = 1             # 0=everyday, 1=Sunday, ..., 7=Saturday
  scheduled_scan_time  = 120           # minutes from midnight
  scheduled_scan_type  = "full"        # quick, full
}
```

### BitLocker

```hcl
module "bitlocker" {
  source    = "MANCHTOOLS/tofutune//modules/bitlocker"
  policy_id = intune_settings_catalog_policy.example.id

  require_device_encryption         = true
  os_drive_encryption_method        = "xts_aes256"  # aes128, aes256, xts_aes128, xts_aes256
  fixed_drive_encryption_method     = "xts_aes256"
  removable_drive_encryption_method = "aes128"
  require_tpm                       = true
  recovery_password_rotation        = "azure_ad_only"  # disabled, azure_ad_only, azure_ad_and_hybrid
  backup_recovery_to_azure_ad       = true
}
```

### Windows Update

```hcl
module "windows_update" {
  source    = "MANCHTOOLS/tofutune//modules/windows_update"
  policy_id = intune_settings_catalog_policy.example.id

  automatic_update_mode        = "auto_install_and_reboot_at_scheduled_time"
  feature_update_deferral_days = 14
  quality_update_deferral_days = 3
  active_hours_start           = 8
  active_hours_end             = 18
  feature_update_deadline_days = 7
  quality_update_deadline_days = 3
  grace_period_days            = 2
  allow_driver_updates         = true
}
```

### Windows AI (Copilot, Recall)

```hcl
module "windows_ai" {
  source    = "MANCHTOOLS/tofutune//modules/windows_ai"
  policy_id = intune_settings_catalog_policy.example.id

  # Disable Windows Copilot
  disable_copilot        = true
  disable_copilot_device = true

  # Disable Windows Recall (Windows 11 24H2+)
  disable_recall           = true
  disable_recall_snapshots = true

  # Disable other AI features
  disable_image_creator    = true
  disable_cocreator        = true

  # Disable AI-powered search
  disable_search_highlights = true
  disable_cloud_search      = true

  # Disable tips and suggestions
  disable_tips_notifications   = true
  disable_spotlight            = true
  disable_settings_suggestions = true
}
```

### Privacy and Telemetry

```hcl
module "privacy" {
  source    = "MANCHTOOLS/tofutune//modules/privacy"
  policy_id = intune_settings_catalog_policy.example.id

  # Telemetry level (security, basic, enhanced, full)
  telemetry_level                 = "basic"
  limit_diagnostic_log_collection = true
  limit_dump_collection           = true
  allow_device_name_in_telemetry  = false

  # Disable personalization
  allow_tailored_experiences = false
  disable_advertising_id     = true

  # Activity history
  enable_activity_feed     = false
  publish_user_activities  = false
  upload_user_activities   = false

  # Feedback
  disable_feedback_notifications = true
}
```

### Microsoft Edge

```hcl
module "edge" {
  source    = "MANCHTOOLS/tofutune//modules/edge"
  policy_id = intune_settings_catalog_policy.example.id

  # Homepage and startup
  homepage_url   = "https://intranet.company.com"
  startup_action = "open_new_tab"  # open_new_tab, restore_session, open_urls

  # Security
  enable_smartscreen      = true
  smartscreen_pua_enabled = true
  site_isolation_enabled  = true

  # Privacy
  tracking_prevention_mode = "strict"  # off, basic, balanced, strict
  do_not_track_enabled     = true

  # Features
  password_manager_enabled = true
  disable_copilot_sidebar  = true
  sync_disabled            = false
  block_popups             = true
}
```

### OneDrive

```hcl
module "onedrive" {
  source    = "MANCHTOOLS/tofutune//modules/onedrive"
  policy_id = intune_settings_catalog_policy.example.id

  tenant_id = "00000000-0000-0000-0000-000000000000"

  # Known Folder Move (backup Desktop, Documents, Pictures to OneDrive)
  silently_move_known_folders = true
  prevent_redirect_to_pc      = true

  # Sync settings
  silent_account_config   = true
  files_on_demand_enabled = true

  # Restrictions
  block_personal_account_sync = true

  # Update ring
  update_ring = "production"  # enterprise, production, insiders, deferred
}
```

## Scope Tags

Scope tags allow you to control which Intune objects administrators can see and manage:

```hcl
# Create a scope tag for the engineering team
resource "intune_scope_tag" "engineering" {
  display_name = "Engineering"
  description  = "Scope tag for engineering team devices and policies"
}

# Use the scope tag in a policy
resource "intune_settings_catalog_policy" "engineering_policy" {
  name               = "Engineering Device Configuration"
  platforms          = "windows10AndLater"
  technologies       = "mdm"
  role_scope_tag_ids = [intune_scope_tag.engineering.id]
}
```

### List Existing Scope Tags

```hcl
data "intune_scope_tags" "all" {}

output "all_scope_tags" {
  value = data.intune_scope_tags.all.scope_tags
}
```

## Assignment Filters

Assignment filters allow you to dynamically include or exclude devices from policy assignments based on device properties:

```hcl
# Filter for Surface devices
resource "intune_assignment_filter" "surface_devices" {
  display_name = "Surface Devices"
  description  = "Filter for Microsoft Surface devices"
  platform     = "windows10AndLater"
  rule         = "(device.model -startsWith \"Surface\")"
}

# Filter by manufacturer
resource "intune_assignment_filter" "dell_devices" {
  display_name = "Dell Devices"
  platform     = "windows10AndLater"
  rule         = "(device.manufacturer -eq \"Dell Inc.\")"
}

# Use filter in policy assignment
resource "intune_settings_catalog_policy" "surface_settings" {
  name         = "Surface Pro Settings"
  platforms    = "windows10AndLater"
  technologies = "mdm"

  assignment {
    all_devices = true
    filter_id   = intune_assignment_filter.surface_devices.id
    filter_type = "include"
  }
}
```

### Filter Rule Syntax

| Operator | Description |
|----------|-------------|
| `-eq` | Equals |
| `-ne` | Not equals |
| `-startsWith` | Starts with |
| `-contains` | Contains |
| `-in` | In array |
| `-notIn` | Not in array |

Common device properties for filtering:

- `device.deviceOwnership` - Corporate, Personal
- `device.manufacturer` - Device manufacturer
- `device.model` - Device model
- `device.osVersion` - Operating system version
- `device.deviceCategory` - Device category

### List Existing Filters

```hcl
# List all filters
data "intune_assignment_filters" "all" {}

# Filter by platform
data "intune_assignment_filters" "windows" {
  platform = "windows10AndLater"
}
```

## Modular Policy Design

The provider is designed for modularity. A single Settings Catalog policy can contain settings from multiple modules:

```hcl
# Create base policy
resource "intune_settings_catalog_policy" "windows_baseline" {
  name         = "Corporate Windows Baseline"
  platforms    = "windows10AndLater"
  technologies = "mdm"
}

# Team A manages Defender settings
module "defender" {
  source    = "./modules/defender"
  policy_id = intune_settings_catalog_policy.windows_baseline.id
  # ...
}

# Team B manages BitLocker settings
module "bitlocker" {
  source    = "./modules/bitlocker"
  policy_id = intune_settings_catalog_policy.windows_baseline.id
  # ...
}

# Team C manages Windows Update settings
module "windows_update" {
  source    = "./modules/windows_update"
  policy_id = intune_settings_catalog_policy.windows_baseline.id
  # ...
}
```

## Examples

See the [examples](./examples) directory for complete working examples:

- [Simple Policy](./examples/simple-policy) - Basic Settings Catalog policy
- [Complete Windows Security](./examples/complete-windows-security) - Full security baseline with modules

## Development

### Building

```bash
go build -o terraform-provider-intune
```

### Testing

```bash
go test ./...
```

### Local Development

Create a `~/.terraformrc` file:

```hcl
provider_installation {
  dev_overrides {
    "MANCHTOOLS/tofutune" = "/path/to/your/provider/binary"
  }
  direct {}
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

Most of the code in this repository was generated with the assistance of AI tools. While the code has been tested and reviewed, users should thoroughly test and validate the provider in their own environments before production use.
