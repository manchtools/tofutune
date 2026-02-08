// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/tofutune/tofutune/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SettingsCatalogPolicyResource{}
var _ resource.ResourceWithImportState = &SettingsCatalogPolicyResource{}

// NewSettingsCatalogPolicyResource creates a new resource instance
func NewSettingsCatalogPolicyResource() resource.Resource {
	return &SettingsCatalogPolicyResource{}
}

// SettingsCatalogPolicyResource defines the resource implementation
type SettingsCatalogPolicyResource struct {
	client *clients.GraphClient
}

// SettingsCatalogPolicyResourceModel describes the resource data model
type SettingsCatalogPolicyResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Platforms            types.String `tfsdk:"platforms"`
	Technologies         types.String `tfsdk:"technologies"`
	RoleScopeTagIds      types.List   `tfsdk:"role_scope_tag_ids"`
	TemplateId           types.String `tfsdk:"template_id"`
	CreatedDateTime      types.String `tfsdk:"created_date_time"`
	LastModifiedDateTime types.String `tfsdk:"last_modified_date_time"`
	SettingCount         types.Int64  `tfsdk:"setting_count"`
}

// Metadata returns the resource type name
func (r *SettingsCatalogPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settings_catalog_policy"
}

// Schema defines the schema for the resource
func (r *SettingsCatalogPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Intune Settings Catalog policy. This resource creates the policy container; " +
			"use intune_settings_catalog_policy_settings to add settings to the policy.",
		MarkdownDescription: `
Manages an Intune Settings Catalog policy.

This resource creates the policy container. Use ` + "`intune_settings_catalog_policy_settings`" + ` to add
settings to the policy in a modular way.

## Example Usage

` + "```hcl" + `
resource "intune_settings_catalog_policy" "windows_security" {
  name         = "Windows Security Baseline"
  description  = "Corporate security settings for Windows devices"
  platforms    = "windows10AndLater"
  technologies = "mdm"
}

# Add settings using a separate resource or module
resource "intune_settings_catalog_policy_settings" "defender" {
  policy_id = intune_settings_catalog_policy.windows_security.id

  setting {
    definition_id = "device_vendor_msft_defender_configuration_disablerealtimemonitoring"
    value_type    = "boolean"
    value         = "false"
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
			"name": schema.StringAttribute{
				Description: "The display name of the policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the policy.",
				Optional:    true,
			},
			"platforms": schema.StringAttribute{
				Description: "The platforms for the policy. Valid values: none, android, iOS, macOS, windows10X, " +
					"windows10, linux, unknownFutureValue, androidEnterprise, aosp.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"none",
						"android",
						"iOS",
						"macOS",
						"windows10X",
						"windows10",
						"windows10AndLater",
						"linux",
						"unknownFutureValue",
						"androidEnterprise",
						"aosp",
					),
				},
			},
			"technologies": schema.StringAttribute{
				Description: "The technologies for the policy. Valid values: none, mdm, windows10XManagement, " +
					"configManager, appleRemoteManagement, microsoftSense, exchangeOnline, mobileApplicationManagement, " +
					"linuxMdm, enrollment, endpointPrivilegeManagement, unknownFutureValue, windowsOsRecovery.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"none",
						"mdm",
						"windows10XManagement",
						"configManager",
						"appleRemoteManagement",
						"microsoftSense",
						"exchangeOnline",
						"mobileApplicationManagement",
						"linuxMdm",
						"enrollment",
						"endpointPrivilegeManagement",
						"unknownFutureValue",
						"windowsOsRecovery",
					),
				},
			},
			"role_scope_tag_ids": schema.ListAttribute{
				Description: "List of scope tag IDs for this policy.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"template_id": schema.StringAttribute{
				Description: "The template ID to base the policy on. This determines which settings are available.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_date_time": schema.StringAttribute{
				Description: "The date and time the policy was created.",
				Computed:    true,
			},
			"last_modified_date_time": schema.StringAttribute{
				Description: "The date and time the policy was last modified.",
				Computed:    true,
			},
			"setting_count": schema.Int64Attribute{
				Description: "The number of settings in the policy.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *SettingsCatalogPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SettingsCatalogPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SettingsCatalogPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Settings Catalog policy", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	// Build the policy object
	policy := &clients.SettingsCatalogPolicy{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Platforms:   data.Platforms.ValueString(),
		Technologies: data.Technologies.ValueString(),
	}

	// Add role scope tag IDs if specified
	if !data.RoleScopeTagIds.IsNull() {
		var tagIds []string
		resp.Diagnostics.Append(data.RoleScopeTagIds.ElementsAs(ctx, &tagIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		policy.RoleScopeTagIds = tagIds
	} else {
		// Default to "0" (Default scope tag)
		policy.RoleScopeTagIds = []string{"0"}
	}

	// Add template reference if specified
	if !data.TemplateId.IsNull() && data.TemplateId.ValueString() != "" {
		policy.TemplateReference = &clients.SettingsCatalogTemplateReference{
			TemplateId: data.TemplateId.ValueString(),
		}
	}

	// Create the policy
	created, err := r.client.CreateSettingsCatalogPolicy(ctx, policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Settings Catalog Policy",
			fmt.Sprintf("Could not create policy: %s", err),
		)
		return
	}

	// Update the model with the created policy data
	data.ID = types.StringValue(created.ID)
	data.CreatedDateTime = types.StringValue(created.CreatedDateTime)
	data.LastModifiedDateTime = types.StringValue(created.LastModifiedDateTime)
	data.SettingCount = types.Int64Value(int64(created.SettingCount))

	tflog.Debug(ctx, "Created Settings Catalog policy", map[string]interface{}{
		"id": created.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *SettingsCatalogPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SettingsCatalogPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Settings Catalog policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Get the policy
	policy, err := r.client.GetSettingsCatalogPolicy(ctx, data.ID.ValueString())
	if err != nil {
		// Check if policy was deleted
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Settings Catalog Policy",
			fmt.Sprintf("Could not read policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update the model
	data.Name = types.StringValue(policy.Name)
	data.Description = types.StringValue(policy.Description)
	data.Platforms = types.StringValue(policy.Platforms)
	data.Technologies = types.StringValue(policy.Technologies)
	data.CreatedDateTime = types.StringValue(policy.CreatedDateTime)
	data.LastModifiedDateTime = types.StringValue(policy.LastModifiedDateTime)
	data.SettingCount = types.Int64Value(int64(policy.SettingCount))

	// Handle role scope tag IDs
	if len(policy.RoleScopeTagIds) > 0 {
		tagIds, diags := types.ListValueFrom(ctx, types.StringType, policy.RoleScopeTagIds)
		resp.Diagnostics.Append(diags...)
		data.RoleScopeTagIds = tagIds
	}

	// Handle template reference
	if policy.TemplateReference != nil && policy.TemplateReference.TemplateId != "" {
		data.TemplateId = types.StringValue(policy.TemplateReference.TemplateId)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state
func (r *SettingsCatalogPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SettingsCatalogPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Settings Catalog policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Build the policy update object
	policy := &clients.SettingsCatalogPolicy{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	// Add role scope tag IDs if specified
	if !data.RoleScopeTagIds.IsNull() {
		var tagIds []string
		resp.Diagnostics.Append(data.RoleScopeTagIds.ElementsAs(ctx, &tagIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		policy.RoleScopeTagIds = tagIds
	}

	// Update the policy
	updated, err := r.client.UpdateSettingsCatalogPolicy(ctx, data.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Settings Catalog Policy",
			fmt.Sprintf("Could not update policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update the model with the updated policy data
	data.LastModifiedDateTime = types.StringValue(updated.LastModifiedDateTime)
	data.SettingCount = types.Int64Value(int64(updated.SettingCount))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *SettingsCatalogPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SettingsCatalogPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Settings Catalog policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	err := r.client.DeleteSettingsCatalogPolicy(ctx, data.ID.ValueString())
	if err != nil {
		// Ignore not found errors during delete
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Settings Catalog Policy",
			fmt.Sprintf("Could not delete policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

// ImportState imports the resource state
func (r *SettingsCatalogPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
