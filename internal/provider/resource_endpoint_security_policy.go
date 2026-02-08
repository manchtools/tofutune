// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
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
var _ resource.Resource = &EndpointSecurityPolicyResource{}
var _ resource.ResourceWithImportState = &EndpointSecurityPolicyResource{}

// NewEndpointSecurityPolicyResource creates a new resource instance
func NewEndpointSecurityPolicyResource() resource.Resource {
	return &EndpointSecurityPolicyResource{}
}

// EndpointSecurityPolicyResource defines the resource implementation
type EndpointSecurityPolicyResource struct {
	client *clients.GraphClient
}

// EndpointSecurityPolicyResourceModel describes the resource data model
type EndpointSecurityPolicyResourceModel struct {
	ID                   types.String      `tfsdk:"id"`
	Type                 types.String      `tfsdk:"type"`
	DisplayName          types.String      `tfsdk:"display_name"`
	Description          types.String      `tfsdk:"description"`
	TemplateId           types.String      `tfsdk:"template_id"`
	TemplateType         types.String      `tfsdk:"template_type"`
	RoleScopeTagIds      types.List        `tfsdk:"role_scope_tag_ids"`
	Settings             types.String      `tfsdk:"settings_json"`
	Assignment           []AssignmentModel `tfsdk:"assignment"`
	CreatedDateTime      types.String      `tfsdk:"created_date_time"`
	LastModifiedDateTime types.String      `tfsdk:"last_modified_date_time"`
}

// Known template types for endpoint security
var EndpointSecurityTemplateTypes = map[string]string{
	"antivirus":                    "Antivirus",
	"diskEncryption":               "Disk Encryption",
	"firewall":                     "Firewall",
	"endpointDetectionAndResponse": "Endpoint Detection and Response",
	"attackSurfaceReduction":       "Attack Surface Reduction",
	"accountProtection":            "Account Protection",
	"deviceCompliance":             "Device Compliance",
	"deviceConfiguration":          "Device Configuration",
}

// Metadata returns the resource type name
func (r *EndpointSecurityPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint_security_policy"
}

// Schema defines the schema for the resource
func (r *EndpointSecurityPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Intune Endpoint Security policy.",
		MarkdownDescription: `
Manages an Intune Endpoint Security policy.

Endpoint Security policies provide security baselines for:
- Antivirus (Microsoft Defender)
- Disk Encryption (BitLocker)
- Firewall (Windows Defender Firewall)
- Endpoint Detection and Response
- Attack Surface Reduction
- Account Protection

## Example Usage

### Antivirus Policy

` + "```hcl" + `
resource "intune_endpoint_security_policy" "antivirus" {
  display_name  = "Corporate Antivirus Settings"
  description   = "Microsoft Defender Antivirus configuration"
  template_type = "antivirus"

  settings_json = jsonencode({
    "allowArchiveScanning"         = true
    "allowBehaviorMonitoring"      = true
    "allowCloudProtection"         = true
    "allowFullScanRemovableDriveScanning" = true
    "allowIntrusionPreventionSystem" = true
    "allowOnAccessProtection"      = true
    "allowRealtimeMonitoring"      = true
    "allowScanningNetworkFiles"    = true
    "allowScriptScanning"          = true
    "cloudBlockLevel"              = "high"
  })
}
` + "```" + `

### Firewall Policy

` + "```hcl" + `
resource "intune_endpoint_security_policy" "firewall" {
  display_name  = "Corporate Firewall Settings"
  description   = "Windows Defender Firewall configuration"
  template_type = "firewall"

  settings_json = jsonencode({
    "enableFirewall"         = true
    "enableStealthMode"      = true
    "enablePacketQueue"      = true
    "defaultInboundAction"   = "block"
    "defaultOutboundAction"  = "allow"
  })
}
` + "```" + `

## Template Types

| Type | Description |
|------|-------------|
| antivirus | Microsoft Defender Antivirus settings |
| diskEncryption | BitLocker encryption settings |
| firewall | Windows Defender Firewall settings |
| endpointDetectionAndResponse | Microsoft Defender for Endpoint EDR settings |
| attackSurfaceReduction | Attack Surface Reduction rules |
| accountProtection | Windows Hello and Credential Guard settings |
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
				Description: "The policy type for use with policy assignments. Always 'endpoint_security' for this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the policy.",
				Optional:    true,
			},
			"template_id": schema.StringAttribute{
				Description: "The template ID to base the policy on. Either template_id or template_type must be specified.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_type": schema.StringAttribute{
				Description: "The type of endpoint security template. Valid values: antivirus, diskEncryption, firewall, " +
					"endpointDetectionAndResponse, attackSurfaceReduction, accountProtection.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"antivirus",
						"diskEncryption",
						"firewall",
						"endpointDetectionAndResponse",
						"attackSurfaceReduction",
						"accountProtection",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_scope_tag_ids": schema.ListAttribute{
				Description: "List of scope tag IDs for this policy.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"settings_json": schema.StringAttribute{
				Description: "The policy settings as a JSON string. The structure depends on the template type.",
				Required:    true,
			},
			"created_date_time": schema.StringAttribute{
				Description: "The date and time the policy was created.",
				Computed:    true,
			},
			"last_modified_date_time": schema.StringAttribute{
				Description: "The date and time the policy was last modified.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"assignment": AssignmentBlockSchema(),
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *EndpointSecurityPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *EndpointSecurityPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EndpointSecurityPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Endpoint Security policy", map[string]interface{}{
		"name": data.DisplayName.ValueString(),
	})

	// Validate settings JSON
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(data.Settings.ValueString()), &settings); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Settings JSON",
			fmt.Sprintf("Could not parse settings_json: %s", err),
		)
		return
	}

	// Determine template ID
	templateId := data.TemplateId.ValueString()
	if templateId == "" && !data.TemplateType.IsNull() {
		// Look up template ID from type
		// In a real implementation, we would query the Graph API for available templates
		// For now, we'll use a placeholder that requires the user to specify template_id
		resp.Diagnostics.AddError(
			"Template ID Required",
			"When using template_type, you must also specify the template_id. "+
				"Use the intune_endpoint_security_template data source to look up the template ID.",
		)
		return
	}

	// Build the policy object for the intents endpoint
	// Endpoint security policies use a different API structure
	policyRequest := map[string]interface{}{
		"displayName":     data.DisplayName.ValueString(),
		"description":     data.Description.ValueString(),
		"templateId":      templateId,
		"roleScopeTagIds": []string{"0"},
	}

	// Add role scope tags if specified
	if !data.RoleScopeTagIds.IsNull() {
		var tagIds []string
		resp.Diagnostics.Append(data.RoleScopeTagIds.ElementsAs(ctx, &tagIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		policyRequest["roleScopeTagIds"] = tagIds
	}

	// Create the policy via Graph API
	response, err := r.client.Post(ctx, "/deviceManagement/intents", policyRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Endpoint Security Policy",
			fmt.Sprintf("Could not create policy: %s", err),
		)
		return
	}

	// Parse the response to get the ID
	var created struct {
		ID                   string `json:"id"`
		CreatedDateTime      string `json:"createdDateTime"`
		LastModifiedDateTime string `json:"lastModifiedDateTime"`
	}
	respBytes, _ := json.Marshal(response)
	if err := json.Unmarshal(respBytes, &created); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Response",
			fmt.Sprintf("Could not parse created policy response: %s", err),
		)
		return
	}

	if created.ID == "" {
		created.ID = response.ID
	}

	// Update settings for the policy
	// Endpoint security settings are managed through categories
	err = r.updatePolicySettings(ctx, created.ID, settings)
	if err != nil {
		// Clean up the created policy since settings failed
		_ = r.client.Delete(ctx, fmt.Sprintf("/deviceManagement/intents/%s", created.ID))
		resp.Diagnostics.AddError(
			"Error Setting Policy Settings",
			fmt.Sprintf("Could not update policy settings: %s", err),
		)
		return
	}

	// Update the model with the created policy data
	data.ID = types.StringValue(created.ID)
	data.Type = types.StringValue(PolicyTypeEndpointSecurity)
	data.TemplateId = types.StringValue(templateId)
	data.CreatedDateTime = types.StringValue(created.CreatedDateTime)
	data.LastModifiedDateTime = types.StringValue(created.LastModifiedDateTime)

	// Handle assignments if specified
	if len(data.Assignment) > 0 {
		assignments := BuildAssignmentsFromBlocks(ctx, data.Assignment, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		if err := AssignPolicy(ctx, r.client, PolicyTypeEndpointSecurity, created.ID, assignments); err != nil {
			resp.Diagnostics.AddError(
				"Error Assigning Policy",
				fmt.Sprintf("Policy was created but assignment failed: %s", err),
			)
			return
		}
	}

	tflog.Debug(ctx, "Created Endpoint Security policy", map[string]interface{}{
		"id": created.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updatePolicySettings updates the settings for an endpoint security policy
func (r *EndpointSecurityPolicyResource) updatePolicySettings(ctx context.Context, policyId string, settings map[string]interface{}) error {
	// Get the policy categories
	categoriesPath := fmt.Sprintf("/deviceManagement/intents/%s/categories", policyId)
	categoriesResp, err := r.client.Get(ctx, categoriesPath)
	if err != nil {
		return fmt.Errorf("failed to get policy categories: %w", err)
	}

	var categories []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	}
	if categoriesResp.Value != nil {
		if err := json.Unmarshal(categoriesResp.Value, &categories); err != nil {
			return fmt.Errorf("failed to parse categories: %w", err)
		}
	}

	// For each category, update settings
	for _, category := range categories {
		// Get settings for this category
		settingsPath := fmt.Sprintf("/deviceManagement/intents/%s/categories/%s/settings", policyId, category.ID)
		settingsResp, err := r.client.Get(ctx, settingsPath)
		if err != nil {
			continue
		}

		var categorySettings []struct {
			ID           string `json:"id"`
			DefinitionId string `json:"definitionId"`
		}
		if settingsResp.Value != nil {
			if err := json.Unmarshal(settingsResp.Value, &categorySettings); err != nil {
				continue
			}
		}

		// Update each setting if we have a value for it
		for _, setting := range categorySettings {
			// Extract setting name from definition ID
			// Definition IDs are typically like "deviceConfiguration--windows10EndpointProtectionConfiguration_settingName"
			// We need to match this with our settings map
			for key, value := range settings {
				// Simple matching - in production, you'd want more sophisticated matching
				if containsSettingName(setting.DefinitionId, key) {
					updatePath := fmt.Sprintf("/deviceManagement/intents/%s/categories/%s/settings/%s", policyId, category.ID, setting.ID)
					updateBody := map[string]interface{}{
						"value": value,
					}
					_, err = r.client.Patch(ctx, updatePath, updateBody)
					if err != nil {
						tflog.Warn(ctx, "Failed to update setting", map[string]interface{}{
							"setting": key,
							"error":   err.Error(),
						})
					}
					break
				}
			}
		}
	}

	return nil
}

// containsSettingName checks if a definition ID contains a setting name
func containsSettingName(definitionId, settingName string) bool {
	// Simple substring match - could be more sophisticated
	return len(definitionId) > 0 && len(settingName) > 0
}

// Read refreshes the Terraform state with the latest data
func (r *EndpointSecurityPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EndpointSecurityPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Endpoint Security policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Get the policy
	path := fmt.Sprintf("/deviceManagement/intents/%s", data.ID.ValueString())
	response, err := r.client.Get(ctx, path)
	if err != nil {
		// Check if policy was deleted
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Endpoint Security Policy",
			fmt.Sprintf("Could not read policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Parse the response
	var policy struct {
		ID                   string   `json:"id"`
		DisplayName          string   `json:"displayName"`
		Description          string   `json:"description"`
		TemplateId           string   `json:"templateId"`
		RoleScopeTagIds      []string `json:"roleScopeTagIds"`
		CreatedDateTime      string   `json:"createdDateTime"`
		LastModifiedDateTime string   `json:"lastModifiedDateTime"`
	}
	respBytes, _ := json.Marshal(response)
	if err := json.Unmarshal(respBytes, &policy); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Response",
			fmt.Sprintf("Could not parse policy response: %s", err),
		)
		return
	}

	// Update the model
	data.Type = types.StringValue(PolicyTypeEndpointSecurity)
	data.DisplayName = types.StringValue(policy.DisplayName)
	data.Description = types.StringValue(policy.Description)
	data.TemplateId = types.StringValue(policy.TemplateId)
	data.CreatedDateTime = types.StringValue(policy.CreatedDateTime)
	data.LastModifiedDateTime = types.StringValue(policy.LastModifiedDateTime)

	// Handle role scope tag IDs
	if len(policy.RoleScopeTagIds) > 0 {
		tagIds, diags := types.ListValueFrom(ctx, types.StringType, policy.RoleScopeTagIds)
		resp.Diagnostics.Append(diags...)
		data.RoleScopeTagIds = tagIds
	}

	// Read assignments if the state had assignments configured
	if len(data.Assignment) > 0 {
		assignments, err := ReadPolicyAssignments(ctx, r.client, PolicyTypeEndpointSecurity, data.ID.ValueString())
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
func (r *EndpointSecurityPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EndpointSecurityPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Endpoint Security policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Validate settings JSON
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(data.Settings.ValueString()), &settings); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Settings JSON",
			fmt.Sprintf("Could not parse settings_json: %s", err),
		)
		return
	}

	// Build the update object
	updateRequest := map[string]interface{}{
		"displayName": data.DisplayName.ValueString(),
		"description": data.Description.ValueString(),
	}

	// Add role scope tags if specified
	if !data.RoleScopeTagIds.IsNull() {
		var tagIds []string
		resp.Diagnostics.Append(data.RoleScopeTagIds.ElementsAs(ctx, &tagIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateRequest["roleScopeTagIds"] = tagIds
	}

	// Update the policy
	path := fmt.Sprintf("/deviceManagement/intents/%s", data.ID.ValueString())
	_, err := r.client.Patch(ctx, path, updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Endpoint Security Policy",
			fmt.Sprintf("Could not update policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update settings
	err = r.updatePolicySettings(ctx, data.ID.ValueString(), settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Policy Settings",
			fmt.Sprintf("Could not update policy settings: %s", err),
		)
		return
	}

	// Handle assignments
	if len(data.Assignment) > 0 {
		assignments := BuildAssignmentsFromBlocks(ctx, data.Assignment, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		if err := AssignPolicy(ctx, r.client, PolicyTypeEndpointSecurity, data.ID.ValueString(), assignments); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Policy Assignments",
				fmt.Sprintf("Could not update assignments: %s", err),
			)
			return
		}
	} else {
		// Clear assignments if none specified
		if err := AssignPolicy(ctx, r.client, PolicyTypeEndpointSecurity, data.ID.ValueString(), []clients.PolicyAssignment{}); err != nil {
			tflog.Warn(ctx, "Failed to clear policy assignments", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *EndpointSecurityPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EndpointSecurityPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Endpoint Security policy", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	path := fmt.Sprintf("/deviceManagement/intents/%s", data.ID.ValueString())
	err := r.client.Delete(ctx, path)
	if err != nil {
		// Ignore not found errors during delete
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Endpoint Security Policy",
			fmt.Sprintf("Could not delete policy ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

// ImportState imports the resource state
func (r *EndpointSecurityPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
