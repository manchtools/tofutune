// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/tofutune/tofutune/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &CompliancePolicyResource{}
var _ resource.ResourceWithImportState = &CompliancePolicyResource{}

// NewCompliancePolicyResource creates a new resource instance
func NewCompliancePolicyResource() resource.Resource {
	return &CompliancePolicyResource{}
}

// CompliancePolicyResource defines the resource implementation
type CompliancePolicyResource struct {
	client *clients.GraphClient
}

// CompliancePolicyResourceModel describes the resource data model for Windows 10 compliance
type CompliancePolicyResourceModel struct {
	ID                                  types.String `tfsdk:"id"`
	Type                                types.String `tfsdk:"type"`
	DisplayName                         types.String `tfsdk:"display_name"`
	Description                         types.String `tfsdk:"description"`
	RoleScopeTagIds                     types.List   `tfsdk:"role_scope_tag_ids"`
	CreatedDateTime                     types.String `tfsdk:"created_date_time"`
	LastModifiedDateTime                types.String `tfsdk:"last_modified_date_time"`

	// Password settings
	PasswordRequired                    types.Bool   `tfsdk:"password_required"`
	PasswordBlockSimple                 types.Bool   `tfsdk:"password_block_simple"`
	PasswordRequiredToUnlockFromIdle    types.Bool   `tfsdk:"password_required_to_unlock_from_idle"`
	PasswordMinutesOfInactivityBeforeLock types.Int64 `tfsdk:"password_minutes_of_inactivity_before_lock"`
	PasswordExpirationDays              types.Int64  `tfsdk:"password_expiration_days"`
	PasswordMinimumLength               types.Int64  `tfsdk:"password_minimum_length"`
	PasswordMinimumCharacterSetCount    types.Int64  `tfsdk:"password_minimum_character_set_count"`
	PasswordRequiredType                types.String `tfsdk:"password_required_type"`
	PasswordPreviousPasswordBlockCount  types.Int64  `tfsdk:"password_previous_password_block_count"`

	// OS version settings
	OsMinimumVersion                    types.String `tfsdk:"os_minimum_version"`
	OsMaximumVersion                    types.String `tfsdk:"os_maximum_version"`
	MobileOsMinimumVersion              types.String `tfsdk:"mobile_os_minimum_version"`
	MobileOsMaximumVersion              types.String `tfsdk:"mobile_os_maximum_version"`

	// Security settings
	RequireHealthyDeviceReport          types.Bool   `tfsdk:"require_healthy_device_report"`
	EarlyLaunchAntiMalwareDriverEnabled types.Bool   `tfsdk:"early_launch_anti_malware_driver_enabled"`
	BitLockerEnabled                    types.Bool   `tfsdk:"bitlocker_enabled"`
	SecureBootEnabled                   types.Bool   `tfsdk:"secure_boot_enabled"`
	CodeIntegrityEnabled                types.Bool   `tfsdk:"code_integrity_enabled"`
	StorageRequireEncryption            types.Bool   `tfsdk:"storage_require_encryption"`
	TpmRequired                         types.Bool   `tfsdk:"tpm_required"`

	// Firewall & Defender settings
	ActiveFirewallRequired              types.Bool   `tfsdk:"active_firewall_required"`
	DefenderEnabled                     types.Bool   `tfsdk:"defender_enabled"`
	DefenderVersion                     types.String `tfsdk:"defender_version"`
	SignatureOutOfDate                  types.Bool   `tfsdk:"signature_out_of_date"`
	RtpEnabled                          types.Bool   `tfsdk:"rtp_enabled"`
	AntivirusRequired                   types.Bool   `tfsdk:"antivirus_required"`
	AntiSpywareRequired                 types.Bool   `tfsdk:"anti_spyware_required"`

	// Threat protection
	DeviceThreatProtectionEnabled       types.Bool   `tfsdk:"device_threat_protection_enabled"`
	DeviceThreatProtectionRequiredSecurityLevel types.String `tfsdk:"device_threat_protection_required_security_level"`

	// Configuration Manager
	ConfigurationManagerComplianceRequired types.Bool `tfsdk:"configuration_manager_compliance_required"`

	// Assignment
	Assignment []AssignmentModel `tfsdk:"assignment"`

	// Scheduled actions
	ScheduledActionsForRule             types.List   `tfsdk:"scheduled_actions_for_rule"`
}

// ScheduledActionModel represents scheduled action configuration
type ScheduledActionModel struct {
	RuleName                       types.String `tfsdk:"rule_name"`
	ScheduledActionConfigurations  types.List   `tfsdk:"scheduled_action_configurations"`
}

// ScheduledActionConfigurationModel represents action configuration
type ScheduledActionConfigurationModel struct {
	ActionType             types.String `tfsdk:"action_type"`
	GracePeriodHours       types.Int64  `tfsdk:"grace_period_hours"`
	NotificationTemplateId types.String `tfsdk:"notification_template_id"`
}

// Metadata returns the resource type name
func (r *CompliancePolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compliance_policy"
}

// Schema defines the schema for the resource
func (r *CompliancePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Windows 10/11 device compliance policy in Microsoft Intune.",
		MarkdownDescription: `
Manages a Windows 10/11 device compliance policy in Microsoft Intune.

Compliance policies define the rules and settings that devices must meet to be considered compliant.
Non-compliant devices can be blocked from accessing corporate resources.

## Example Usage

` + "```hcl" + `
resource "intune_compliance_policy" "windows" {
  display_name = "Windows 10 Compliance Policy"
  description  = "Corporate compliance requirements for Windows devices"

  # Password requirements
  password_required        = true
  password_minimum_length  = 8
  password_required_type   = "alphanumeric"
  password_block_simple    = true

  # Security requirements
  bitlocker_enabled        = true
  secure_boot_enabled      = true
  code_integrity_enabled   = true
  tpm_required             = true

  # Defender requirements
  defender_enabled         = true
  rtp_enabled              = true
  antivirus_required       = true
  anti_spyware_required    = true

  # Firewall
  active_firewall_required = true

  # OS version
  os_minimum_version       = "10.0.19041.0"

  # Actions for non-compliance
  scheduled_actions_for_rule {
    rule_name = "DeviceNotCompliant"
    scheduled_action_configurations {
      action_type        = "block"
      grace_period_hours = 24
    }
  }
}
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The policy type for use with intune_policy_assignment. Always 'compliance'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the compliance policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the compliance policy.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"role_scope_tag_ids": schema.ListAttribute{
				Description: "List of scope tag IDs for this policy.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"created_date_time": schema.StringAttribute{
				Description: "The date and time the policy was created.",
				Computed:    true,
			},
			"last_modified_date_time": schema.StringAttribute{
				Description: "The date and time the policy was last modified.",
				Computed:    true,
			},

			// Password settings
			"password_required": schema.BoolAttribute{
				Description: "Require a password to unlock the device.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"password_block_simple": schema.BoolAttribute{
				Description: "Block simple passwords like 1234 or 1111.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"password_required_to_unlock_from_idle": schema.BoolAttribute{
				Description: "Require a password to unlock an idle device.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"password_minutes_of_inactivity_before_lock": schema.Int64Attribute{
				Description: "Minutes of inactivity before password is required.",
				Optional:    true,
			},
			"password_expiration_days": schema.Int64Attribute{
				Description: "Number of days until the password expires.",
				Optional:    true,
			},
			"password_minimum_length": schema.Int64Attribute{
				Description: "Minimum password length.",
				Optional:    true,
			},
			"password_minimum_character_set_count": schema.Int64Attribute{
				Description: "Minimum number of character sets required in password.",
				Optional:    true,
			},
			"password_required_type": schema.StringAttribute{
				Description: "Type of password required. Valid values: deviceDefault, alphanumeric, numeric.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("deviceDefault"),
				Validators: []validator.String{
					stringvalidator.OneOf("deviceDefault", "alphanumeric", "numeric"),
				},
			},
			"password_previous_password_block_count": schema.Int64Attribute{
				Description: "Number of previous passwords to block.",
				Optional:    true,
			},

			// OS version settings
			"os_minimum_version": schema.StringAttribute{
				Description: "Minimum OS version required.",
				Optional:    true,
			},
			"os_maximum_version": schema.StringAttribute{
				Description: "Maximum OS version allowed.",
				Optional:    true,
			},
			"mobile_os_minimum_version": schema.StringAttribute{
				Description: "Minimum mobile OS version required.",
				Optional:    true,
			},
			"mobile_os_maximum_version": schema.StringAttribute{
				Description: "Maximum mobile OS version allowed.",
				Optional:    true,
			},

			// Security settings
			"require_healthy_device_report": schema.BoolAttribute{
				Description: "Require devices to report healthy to Windows Device Health Attestation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"early_launch_anti_malware_driver_enabled": schema.BoolAttribute{
				Description: "Require early launch anti-malware driver to be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"bitlocker_enabled": schema.BoolAttribute{
				Description: "Require BitLocker to be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"secure_boot_enabled": schema.BoolAttribute{
				Description: "Require Secure Boot to be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"code_integrity_enabled": schema.BoolAttribute{
				Description: "Require code integrity to be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"storage_require_encryption": schema.BoolAttribute{
				Description: "Require encryption on the device.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"tpm_required": schema.BoolAttribute{
				Description: "Require Trusted Platform Module (TPM).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			// Firewall & Defender settings
			"active_firewall_required": schema.BoolAttribute{
				Description: "Require Windows Firewall to be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"defender_enabled": schema.BoolAttribute{
				Description: "Require Windows Defender to be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"defender_version": schema.StringAttribute{
				Description: "Minimum Windows Defender version required.",
				Optional:    true,
			},
			"signature_out_of_date": schema.BoolAttribute{
				Description: "Require Windows Defender signatures to be up to date.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"rtp_enabled": schema.BoolAttribute{
				Description: "Require Windows Defender real-time protection to be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"antivirus_required": schema.BoolAttribute{
				Description: "Require antivirus to be registered and monitoring.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"anti_spyware_required": schema.BoolAttribute{
				Description: "Require anti-spyware to be registered and monitoring.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			// Threat protection
			"device_threat_protection_enabled": schema.BoolAttribute{
				Description: "Require device threat protection.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"device_threat_protection_required_security_level": schema.StringAttribute{
				Description: "Required security level for device threat protection. Valid values: unavailable, secured, low, medium, high, notSet.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("notSet"),
				Validators: []validator.String{
					stringvalidator.OneOf("unavailable", "secured", "low", "medium", "high", "notSet"),
				},
			},

			// Configuration Manager
			"configuration_manager_compliance_required": schema.BoolAttribute{
				Description: "Require Configuration Manager compliance.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"assignment": AssignmentBlockSchema(),
			"scheduled_actions_for_rule": schema.ListNestedBlock{
				Description: "Scheduled actions for non-compliance.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"rule_name": schema.StringAttribute{
							Description: "The rule name. Use 'DeviceNotCompliant' for the default rule.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("DeviceNotCompliant"),
						},
					},
					Blocks: map[string]schema.Block{
						"scheduled_action_configurations": schema.ListNestedBlock{
							Description: "Action configurations for non-compliance.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"action_type": schema.StringAttribute{
										Description: "The action type. Valid values: block, retire, wipe, removeResourceAccess, pushNotification.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("block", "retire", "wipe", "removeResourceAccessOutsideResource", "pushNotification"),
										},
									},
									"grace_period_hours": schema.Int64Attribute{
										Description: "Number of hours before the action is enforced. 0 for immediate.",
										Optional:    true,
										Computed:    true,
										Default:     int64default.StaticInt64(0),
									},
									"notification_template_id": schema.StringAttribute{
										Description: "The notification template ID to use.",
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *CompliancePolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = providerData.GraphClient
}

// Create creates the resource and sets the initial Terraform state
func (r *CompliancePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CompliancePolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Compliance policy", map[string]interface{}{
		"name": data.DisplayName.ValueString(),
	})

	// Build the policy object
	policy := r.buildPolicy(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the policy
	created, err := r.client.CreateCompliancePolicy(ctx, policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Compliance Policy",
			fmt.Sprintf("Could not create policy: %s", err),
		)
		return
	}

	// Update the model with the created policy data
	data.ID = types.StringValue(created.ID)
	data.Type = types.StringValue(PolicyTypeCompliance)
	data.CreatedDateTime = types.StringValue(created.CreatedDateTime)
	data.LastModifiedDateTime = types.StringValue(created.LastModifiedDateTime)

	// Handle assignments if specified
	if len(data.Assignment) > 0 {
		assignments := BuildAssignmentsFromBlocks(ctx, data.Assignment, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		if err := AssignPolicy(ctx, r.client, PolicyTypeCompliance, created.ID, assignments); err != nil {
			resp.Diagnostics.AddError(
				"Error Assigning Policy",
				fmt.Sprintf("Policy was created but assignment failed: %s", err),
			)
			return
		}
	}

	tflog.Debug(ctx, "Created Compliance policy", map[string]interface{}{
		"id": created.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *CompliancePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CompliancePolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Compliance policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Get the policy
	policy, err := r.client.GetCompliancePolicy(ctx, data.ID.ValueString())
	if err != nil {
		// Check if policy was deleted
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Compliance Policy",
			fmt.Sprintf("Could not read policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update the model
	r.updateModel(&data, policy, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read assignments if the state had assignments configured
	if len(data.Assignment) > 0 {
		assignments, err := ReadPolicyAssignments(ctx, r.client, PolicyTypeCompliance, data.ID.ValueString())
		if err != nil {
			tflog.Warn(ctx, "Failed to read policy assignments", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			data.Assignment = assignments
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state
func (r *CompliancePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CompliancePolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Compliance policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Build the policy update object
	policy := r.buildPolicy(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the policy
	updated, err := r.client.UpdateCompliancePolicy(ctx, data.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Compliance Policy",
			fmt.Sprintf("Could not update policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update the model with the updated policy data
	data.LastModifiedDateTime = types.StringValue(updated.LastModifiedDateTime)

	// Handle assignments
	if len(data.Assignment) > 0 {
		assignments := BuildAssignmentsFromBlocks(ctx, data.Assignment, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		if err := AssignPolicy(ctx, r.client, PolicyTypeCompliance, data.ID.ValueString(), assignments); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Policy Assignments",
				fmt.Sprintf("Could not update assignments: %s", err),
			)
			return
		}
	} else {
		// Clear assignments if none specified
		if err := AssignPolicy(ctx, r.client, PolicyTypeCompliance, data.ID.ValueString(), []clients.PolicyAssignment{}); err != nil {
			tflog.Warn(ctx, "Failed to clear policy assignments", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *CompliancePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CompliancePolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Compliance policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	err := r.client.DeleteCompliancePolicy(ctx, data.ID.ValueString())
	if err != nil {
		// Ignore not found errors during delete
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Compliance Policy",
			fmt.Sprintf("Could not delete policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

// ImportState imports the resource state
func (r *CompliancePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildPolicy builds the API policy object from the Terraform model
func (r *CompliancePolicyResource) buildPolicy(data *CompliancePolicyResourceModel, diags *diag.Diagnostics) *clients.CompliancePolicy {
	policy := &clients.CompliancePolicy{
		ODataType:   "#microsoft.graph.windows10CompliancePolicy",
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),

		// Password settings
		PasswordRequired:                 data.PasswordRequired.ValueBool(),
		PasswordBlockSimple:              data.PasswordBlockSimple.ValueBool(),
		PasswordRequiredToUnlockFromIdle: data.PasswordRequiredToUnlockFromIdle.ValueBool(),
		PasswordRequiredType:             data.PasswordRequiredType.ValueString(),

		// Security settings
		RequireHealthyDeviceReport:          data.RequireHealthyDeviceReport.ValueBool(),
		EarlyLaunchAntiMalwareDriverEnabled: data.EarlyLaunchAntiMalwareDriverEnabled.ValueBool(),
		BitLockerEnabled:                    data.BitLockerEnabled.ValueBool(),
		SecureBootEnabled:                   data.SecureBootEnabled.ValueBool(),
		CodeIntegrityEnabled:                data.CodeIntegrityEnabled.ValueBool(),
		StorageRequireEncryption:            data.StorageRequireEncryption.ValueBool(),
		TpmRequired:                         data.TpmRequired.ValueBool(),

		// Firewall & Defender
		ActiveFirewallRequired: data.ActiveFirewallRequired.ValueBool(),
		DefenderEnabled:        data.DefenderEnabled.ValueBool(),
		SignatureOutOfDate:     data.SignatureOutOfDate.ValueBool(),
		RtpEnabled:             data.RtpEnabled.ValueBool(),
		AntivirusRequired:      data.AntivirusRequired.ValueBool(),
		AntiSpywareRequired:    data.AntiSpywareRequired.ValueBool(),

		// Threat protection
		DeviceThreatProtectionEnabled:               data.DeviceThreatProtectionEnabled.ValueBool(),
		DeviceThreatProtectionRequiredSecurityLevel: data.DeviceThreatProtectionRequiredSecurityLevel.ValueString(),

		// Configuration Manager
		ConfigurationManagerComplianceRequired: data.ConfigurationManagerComplianceRequired.ValueBool(),
	}

	// Optional integer fields
	if !data.PasswordMinutesOfInactivityBeforeLock.IsNull() {
		val := int(data.PasswordMinutesOfInactivityBeforeLock.ValueInt64())
		policy.PasswordMinutesOfInactivityBeforeLock = &val
	}
	if !data.PasswordExpirationDays.IsNull() {
		val := int(data.PasswordExpirationDays.ValueInt64())
		policy.PasswordExpirationDays = &val
	}
	if !data.PasswordMinimumLength.IsNull() {
		val := int(data.PasswordMinimumLength.ValueInt64())
		policy.PasswordMinimumLength = &val
	}
	if !data.PasswordMinimumCharacterSetCount.IsNull() {
		val := int(data.PasswordMinimumCharacterSetCount.ValueInt64())
		policy.PasswordMinimumCharacterSetCount = &val
	}
	if !data.PasswordPreviousPasswordBlockCount.IsNull() {
		val := int(data.PasswordPreviousPasswordBlockCount.ValueInt64())
		policy.PasswordPreviousPasswordBlockCount = &val
	}

	// Optional string fields
	if !data.OsMinimumVersion.IsNull() {
		policy.OsMinimumVersion = data.OsMinimumVersion.ValueString()
	}
	if !data.OsMaximumVersion.IsNull() {
		policy.OsMaximumVersion = data.OsMaximumVersion.ValueString()
	}
	if !data.MobileOsMinimumVersion.IsNull() {
		policy.MobileOsMinimumVersion = data.MobileOsMinimumVersion.ValueString()
	}
	if !data.MobileOsMaximumVersion.IsNull() {
		policy.MobileOsMaximumVersion = data.MobileOsMaximumVersion.ValueString()
	}
	if !data.DefenderVersion.IsNull() {
		policy.DefenderVersion = data.DefenderVersion.ValueString()
	}

	// Role scope tags
	if !data.RoleScopeTagIds.IsNull() {
		var tagIds []string
		diags.Append(data.RoleScopeTagIds.ElementsAs(context.Background(), &tagIds, false)...)
		policy.RoleScopeTagIds = tagIds
	} else {
		policy.RoleScopeTagIds = []string{"0"}
	}

	// Scheduled actions - default to marking device non-compliant immediately if not specified
	policy.ScheduledActionsForRule = []clients.ComplianceScheduledAction{
		{
			RuleName: "DeviceNotCompliant",
			ScheduledActionConfigurations: []clients.ScheduledActionConfiguration{
				{
					ActionType:       "block",
					GracePeriodHours: 0,
				},
			},
		},
	}

	return policy
}

// updateModel updates the Terraform model from the API policy
func (r *CompliancePolicyResource) updateModel(data *CompliancePolicyResourceModel, policy *clients.CompliancePolicy, diags *diag.Diagnostics) {
	data.DisplayName = types.StringValue(policy.DisplayName)
	data.Type = types.StringValue(PolicyTypeCompliance)
	data.Description = types.StringValue(policy.Description)
	data.CreatedDateTime = types.StringValue(policy.CreatedDateTime)
	data.LastModifiedDateTime = types.StringValue(policy.LastModifiedDateTime)

	// Password settings
	data.PasswordRequired = types.BoolValue(policy.PasswordRequired)
	data.PasswordBlockSimple = types.BoolValue(policy.PasswordBlockSimple)
	data.PasswordRequiredToUnlockFromIdle = types.BoolValue(policy.PasswordRequiredToUnlockFromIdle)
	data.PasswordRequiredType = types.StringValue(policy.PasswordRequiredType)

	// Security settings
	data.RequireHealthyDeviceReport = types.BoolValue(policy.RequireHealthyDeviceReport)
	data.EarlyLaunchAntiMalwareDriverEnabled = types.BoolValue(policy.EarlyLaunchAntiMalwareDriverEnabled)
	data.BitLockerEnabled = types.BoolValue(policy.BitLockerEnabled)
	data.SecureBootEnabled = types.BoolValue(policy.SecureBootEnabled)
	data.CodeIntegrityEnabled = types.BoolValue(policy.CodeIntegrityEnabled)
	data.StorageRequireEncryption = types.BoolValue(policy.StorageRequireEncryption)
	data.TpmRequired = types.BoolValue(policy.TpmRequired)

	// Firewall & Defender
	data.ActiveFirewallRequired = types.BoolValue(policy.ActiveFirewallRequired)
	data.DefenderEnabled = types.BoolValue(policy.DefenderEnabled)
	data.SignatureOutOfDate = types.BoolValue(policy.SignatureOutOfDate)
	data.RtpEnabled = types.BoolValue(policy.RtpEnabled)
	data.AntivirusRequired = types.BoolValue(policy.AntivirusRequired)
	data.AntiSpywareRequired = types.BoolValue(policy.AntiSpywareRequired)

	// Threat protection
	data.DeviceThreatProtectionEnabled = types.BoolValue(policy.DeviceThreatProtectionEnabled)
	data.DeviceThreatProtectionRequiredSecurityLevel = types.StringValue(policy.DeviceThreatProtectionRequiredSecurityLevel)

	// Configuration Manager
	data.ConfigurationManagerComplianceRequired = types.BoolValue(policy.ConfigurationManagerComplianceRequired)

	// Optional integer fields
	if policy.PasswordMinutesOfInactivityBeforeLock != nil {
		data.PasswordMinutesOfInactivityBeforeLock = types.Int64Value(int64(*policy.PasswordMinutesOfInactivityBeforeLock))
	}
	if policy.PasswordExpirationDays != nil {
		data.PasswordExpirationDays = types.Int64Value(int64(*policy.PasswordExpirationDays))
	}
	if policy.PasswordMinimumLength != nil {
		data.PasswordMinimumLength = types.Int64Value(int64(*policy.PasswordMinimumLength))
	}
	if policy.PasswordMinimumCharacterSetCount != nil {
		data.PasswordMinimumCharacterSetCount = types.Int64Value(int64(*policy.PasswordMinimumCharacterSetCount))
	}
	if policy.PasswordPreviousPasswordBlockCount != nil {
		data.PasswordPreviousPasswordBlockCount = types.Int64Value(int64(*policy.PasswordPreviousPasswordBlockCount))
	}

	// Optional string fields
	if policy.OsMinimumVersion != "" {
		data.OsMinimumVersion = types.StringValue(policy.OsMinimumVersion)
	}
	if policy.OsMaximumVersion != "" {
		data.OsMaximumVersion = types.StringValue(policy.OsMaximumVersion)
	}
	if policy.MobileOsMinimumVersion != "" {
		data.MobileOsMinimumVersion = types.StringValue(policy.MobileOsMinimumVersion)
	}
	if policy.MobileOsMaximumVersion != "" {
		data.MobileOsMaximumVersion = types.StringValue(policy.MobileOsMaximumVersion)
	}
	if policy.DefenderVersion != "" {
		data.DefenderVersion = types.StringValue(policy.DefenderVersion)
	}

	// Role scope tags
	if len(policy.RoleScopeTagIds) > 0 {
		tagIds, d := types.ListValueFrom(context.Background(), types.StringType, policy.RoleScopeTagIds)
		diags.Append(d...)
		data.RoleScopeTagIds = tagIds
	}
}
