// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
var _ resource.Resource = &PolicyAssignmentResource{}
var _ resource.ResourceWithImportState = &PolicyAssignmentResource{}

// NewPolicyAssignmentResource creates a new resource instance
func NewPolicyAssignmentResource() resource.Resource {
	return &PolicyAssignmentResource{}
}

// PolicyAssignmentResource defines the resource implementation
type PolicyAssignmentResource struct {
	client *clients.GraphClient
}

// PolicyAssignmentResourceModel describes the resource data model
type PolicyAssignmentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	PolicyID      types.String `tfsdk:"policy_id"`
	PolicyType    types.String `tfsdk:"policy_type"`
	IncludeGroups types.List   `tfsdk:"include_groups"`
	ExcludeGroups types.List   `tfsdk:"exclude_groups"`
	AllDevices    types.Bool   `tfsdk:"all_devices"`
	AllUsers      types.Bool   `tfsdk:"all_users"`
	FilterID      types.String `tfsdk:"filter_id"`
	FilterType    types.String `tfsdk:"filter_type"`
}

// PolicyType constants
const (
	PolicyTypeSettingsCatalog  = "settings_catalog"
	PolicyTypeCompliance       = "compliance"
	PolicyTypeEndpointSecurity = "endpoint_security"
	PolicyTypeDeviceConfig     = "device_configuration"
)

// Metadata returns the resource type name
func (r *PolicyAssignmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_assignment"
}

// Schema defines the schema for the resource
func (r *PolicyAssignmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assigns an Intune policy to Azure AD groups.",
		MarkdownDescription: `
Assigns an Intune policy to Azure AD groups.

This resource manages the assignment of any Intune policy type to groups, supporting
both include and exclude group assignments, as well as assignment filters.

## Example Usage

### Basic Group Assignment

` + "```hcl" + `
resource "intune_policy_assignment" "example" {
  policy_id   = intune_settings_catalog_policy.example.id
  policy_type = "settings_catalog"

  include_groups = [
    data.azuread_group.all_devices.id,
    data.azuread_group.it_department.id,
  ]

  exclude_groups = [
    data.azuread_group.test_devices.id,
  ]
}
` + "```" + `

### Assign to All Devices

` + "```hcl" + `
resource "intune_policy_assignment" "all_devices" {
  policy_id   = intune_compliance_policy.windows.id
  policy_type = "compliance"
  all_devices = true
}
` + "```" + `

### With Assignment Filter

` + "```hcl" + `
resource "intune_policy_assignment" "filtered" {
  policy_id   = intune_settings_catalog_policy.example.id
  policy_type = "settings_catalog"

  include_groups = [data.azuread_group.all_devices.id]

  filter_id   = "00000000-0000-0000-0000-000000000000"
  filter_type = "include"
}
` + "```" + `

## Policy Types

| Type | Description |
|------|-------------|
| settings_catalog | Settings Catalog policies |
| compliance | Device compliance policies |
| endpoint_security | Endpoint security policies |
| device_configuration | Device configuration profiles |
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this assignment (policy_id is used as ID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "The ID of the policy to assign.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_type": schema.StringAttribute{
				Description: "The type of policy. Valid values: settings_catalog, compliance, endpoint_security, device_configuration.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						PolicyTypeSettingsCatalog,
						PolicyTypeCompliance,
						PolicyTypeEndpointSecurity,
						PolicyTypeDeviceConfig,
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"include_groups": schema.ListAttribute{
				Description: "List of Azure AD group IDs to include in the assignment.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"exclude_groups": schema.ListAttribute{
				Description: "List of Azure AD group IDs to exclude from the assignment.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"all_devices": schema.BoolAttribute{
				Description: "Assign to all devices. Cannot be used with include_groups.",
				Optional:    true,
			},
			"all_users": schema.BoolAttribute{
				Description: "Assign to all users. Cannot be used with include_groups.",
				Optional:    true,
			},
			"filter_id": schema.StringAttribute{
				Description: "The ID of the assignment filter to apply.",
				Optional:    true,
			},
			"filter_type": schema.StringAttribute{
				Description: "The type of filter. Valid values: include, exclude.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("include", "exclude"),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *PolicyAssignmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// getAssignmentPath returns the API path for assignments based on policy type
func (r *PolicyAssignmentResource) getAssignmentPath(policyType, policyId string) string {
	switch policyType {
	case PolicyTypeSettingsCatalog:
		return fmt.Sprintf("/deviceManagement/configurationPolicies('%s')/assign", policyId)
	case PolicyTypeCompliance:
		return fmt.Sprintf("/deviceManagement/deviceCompliancePolicies/%s/assign", policyId)
	case PolicyTypeEndpointSecurity:
		return fmt.Sprintf("/deviceManagement/intents/%s/assign", policyId)
	case PolicyTypeDeviceConfig:
		return fmt.Sprintf("/deviceManagement/deviceConfigurations/%s/assign", policyId)
	default:
		return ""
	}
}

// getAssignmentsPath returns the API path for reading assignments based on policy type
func (r *PolicyAssignmentResource) getAssignmentsPath(policyType, policyId string) string {
	switch policyType {
	case PolicyTypeSettingsCatalog:
		return fmt.Sprintf("/deviceManagement/configurationPolicies('%s')/assignments", policyId)
	case PolicyTypeCompliance:
		return fmt.Sprintf("/deviceManagement/deviceCompliancePolicies/%s/assignments", policyId)
	case PolicyTypeEndpointSecurity:
		return fmt.Sprintf("/deviceManagement/intents/%s/assignments", policyId)
	case PolicyTypeDeviceConfig:
		return fmt.Sprintf("/deviceManagement/deviceConfigurations/%s/assignments", policyId)
	default:
		return ""
	}
}

// buildAssignments builds the assignment objects for the API
func (r *PolicyAssignmentResource) buildAssignments(ctx context.Context, data *PolicyAssignmentResourceModel, diags *diag.Diagnostics) []clients.PolicyAssignment {
	var assignments []clients.PolicyAssignment

	// Handle include groups
	if !data.IncludeGroups.IsNull() {
		var groupIds []string
		diags.Append(data.IncludeGroups.ElementsAs(ctx, &groupIds, false)...)
		if diags.HasError() {
			return nil
		}

		for _, groupId := range groupIds {
			target := &clients.AssignmentTarget{
				ODataType: "#microsoft.graph.groupAssignmentTarget",
				GroupId:   groupId,
			}

			// Add filter if specified
			if !data.FilterID.IsNull() && data.FilterID.ValueString() != "" {
				target.DeviceAndAppManagementAssignmentFilterId = data.FilterID.ValueString()
				filterType := "include"
				if !data.FilterType.IsNull() {
					filterType = data.FilterType.ValueString()
				}
				target.DeviceAndAppManagementAssignmentFilterType = filterType
			}

			assignments = append(assignments, clients.PolicyAssignment{
				Target: target,
			})
		}
	}

	// Handle all devices
	if !data.AllDevices.IsNull() && data.AllDevices.ValueBool() {
		target := &clients.AssignmentTarget{
			ODataType: "#microsoft.graph.allDevicesAssignmentTarget",
		}

		// Add filter if specified
		if !data.FilterID.IsNull() && data.FilterID.ValueString() != "" {
			target.DeviceAndAppManagementAssignmentFilterId = data.FilterID.ValueString()
			filterType := "include"
			if !data.FilterType.IsNull() {
				filterType = data.FilterType.ValueString()
			}
			target.DeviceAndAppManagementAssignmentFilterType = filterType
		}

		assignments = append(assignments, clients.PolicyAssignment{
			Target: target,
		})
	}

	// Handle all users
	if !data.AllUsers.IsNull() && data.AllUsers.ValueBool() {
		target := &clients.AssignmentTarget{
			ODataType: "#microsoft.graph.allLicensedUsersAssignmentTarget",
		}

		// Add filter if specified
		if !data.FilterID.IsNull() && data.FilterID.ValueString() != "" {
			target.DeviceAndAppManagementAssignmentFilterId = data.FilterID.ValueString()
			filterType := "include"
			if !data.FilterType.IsNull() {
				filterType = data.FilterType.ValueString()
			}
			target.DeviceAndAppManagementAssignmentFilterType = filterType
		}

		assignments = append(assignments, clients.PolicyAssignment{
			Target: target,
		})
	}

	// Handle exclude groups
	if !data.ExcludeGroups.IsNull() {
		var groupIds []string
		diags.Append(data.ExcludeGroups.ElementsAs(ctx, &groupIds, false)...)
		if diags.HasError() {
			return nil
		}

		for _, groupId := range groupIds {
			assignments = append(assignments, clients.PolicyAssignment{
				Target: &clients.AssignmentTarget{
					ODataType: "#microsoft.graph.exclusionGroupAssignmentTarget",
					GroupId:   groupId,
				},
			})
		}
	}

	return assignments
}

// Create creates the resource and sets the initial Terraform state
func (r *PolicyAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PolicyAssignmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyId := data.PolicyID.ValueString()
	policyType := data.PolicyType.ValueString()

	tflog.Debug(ctx, "Creating policy assignment", map[string]interface{}{
		"policy_id":   policyId,
		"policy_type": policyType,
	})

	// Build assignments
	assignments := r.buildAssignments(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the assignments
	assignPath := r.getAssignmentPath(policyType, policyId)
	if assignPath == "" {
		resp.Diagnostics.AddError(
			"Invalid Policy Type",
			fmt.Sprintf("Unknown policy type: %s", policyType),
		)
		return
	}

	body := map[string]interface{}{
		"assignments": assignments,
	}

	_, err := r.client.Post(ctx, assignPath, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Policy Assignment",
			fmt.Sprintf("Could not create assignment: %s", err),
		)
		return
	}

	// Use policy ID as the resource ID
	data.ID = types.StringValue(policyId)

	tflog.Debug(ctx, "Created policy assignment", map[string]interface{}{
		"policy_id": policyId,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *PolicyAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PolicyAssignmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyId := data.PolicyID.ValueString()
	policyType := data.PolicyType.ValueString()

	tflog.Debug(ctx, "Reading policy assignment", map[string]interface{}{
		"policy_id":   policyId,
		"policy_type": policyType,
	})

	// Get current assignments
	assignmentsPath := r.getAssignmentsPath(policyType, policyId)
	if assignmentsPath == "" {
		resp.Diagnostics.AddError(
			"Invalid Policy Type",
			fmt.Sprintf("Unknown policy type: %s", policyType),
		)
		return
	}

	response, err := r.client.Get(ctx, assignmentsPath)
	if err != nil {
		// Check if policy was deleted
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Policy Assignment",
			fmt.Sprintf("Could not read assignments for policy ID %s: %s", policyId, err),
		)
		return
	}

	// Parse assignments
	var assignments []struct {
		ID     string `json:"id"`
		Target struct {
			ODataType string `json:"@odata.type"`
			GroupId   string `json:"groupId"`
			FilterId  string `json:"deviceAndAppManagementAssignmentFilterId"`
			FilterType string `json:"deviceAndAppManagementAssignmentFilterType"`
		} `json:"target"`
	}

	if response.Value != nil {
		if err := json.Unmarshal(response.Value, &assignments); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Response",
				fmt.Sprintf("Could not parse assignments: %s", err),
			)
			return
		}
	}

	// Extract group IDs
	var includeGroups []string
	var excludeGroups []string
	var allDevices, allUsers bool

	for _, assignment := range assignments {
		switch {
		case strings.Contains(assignment.Target.ODataType, "groupAssignmentTarget"):
			includeGroups = append(includeGroups, assignment.Target.GroupId)
		case strings.Contains(assignment.Target.ODataType, "exclusionGroupAssignmentTarget"):
			excludeGroups = append(excludeGroups, assignment.Target.GroupId)
		case strings.Contains(assignment.Target.ODataType, "allDevicesAssignmentTarget"):
			allDevices = true
		case strings.Contains(assignment.Target.ODataType, "allLicensedUsersAssignmentTarget"):
			allUsers = true
		}

		// Capture filter info from first assignment
		if assignment.Target.FilterId != "" && data.FilterID.IsNull() {
			data.FilterID = types.StringValue(assignment.Target.FilterId)
			data.FilterType = types.StringValue(assignment.Target.FilterType)
		}
	}

	// Update model
	if len(includeGroups) > 0 {
		groupList, diags := types.ListValueFrom(ctx, types.StringType, includeGroups)
		resp.Diagnostics.Append(diags...)
		data.IncludeGroups = groupList
	}
	if len(excludeGroups) > 0 {
		groupList, diags := types.ListValueFrom(ctx, types.StringType, excludeGroups)
		resp.Diagnostics.Append(diags...)
		data.ExcludeGroups = groupList
	}
	if allDevices {
		data.AllDevices = types.BoolValue(true)
	}
	if allUsers {
		data.AllUsers = types.BoolValue(true)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state
func (r *PolicyAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PolicyAssignmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyId := data.PolicyID.ValueString()
	policyType := data.PolicyType.ValueString()

	tflog.Debug(ctx, "Updating policy assignment", map[string]interface{}{
		"policy_id":   policyId,
		"policy_type": policyType,
	})

	// Build new assignments
	assignments := r.buildAssignments(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update assignments (this replaces all assignments)
	assignPath := r.getAssignmentPath(policyType, policyId)
	if assignPath == "" {
		resp.Diagnostics.AddError(
			"Invalid Policy Type",
			fmt.Sprintf("Unknown policy type: %s", policyType),
		)
		return
	}

	body := map[string]interface{}{
		"assignments": assignments,
	}

	_, err := r.client.Post(ctx, assignPath, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Policy Assignment",
			fmt.Sprintf("Could not update assignment: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *PolicyAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PolicyAssignmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyId := data.PolicyID.ValueString()
	policyType := data.PolicyType.ValueString()

	tflog.Debug(ctx, "Deleting policy assignment", map[string]interface{}{
		"policy_id":   policyId,
		"policy_type": policyType,
	})

	// Clear assignments by sending empty array
	assignPath := r.getAssignmentPath(policyType, policyId)
	if assignPath == "" {
		resp.Diagnostics.AddError(
			"Invalid Policy Type",
			fmt.Sprintf("Unknown policy type: %s", policyType),
		)
		return
	}

	body := map[string]interface{}{
		"assignments": []interface{}{},
	}

	_, err := r.client.Post(ctx, assignPath, body)
	if err != nil {
		// Ignore not found errors during delete
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Policy Assignment",
			fmt.Sprintf("Could not delete assignment: %s", err),
		)
		return
	}
}

// ImportState imports the resource state
func (r *PolicyAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: policy_type:policy_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: policy_type:policy_id",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
