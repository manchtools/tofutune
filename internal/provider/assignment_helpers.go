// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/MANCHTOOLS/tofutune/internal/clients"
)

// AssignmentModel represents an inline assignment block
type AssignmentModel struct {
	IncludeGroups types.List   `tfsdk:"include_groups"`
	ExcludeGroups types.List   `tfsdk:"exclude_groups"`
	AllDevices    types.Bool   `tfsdk:"all_devices"`
	AllUsers      types.Bool   `tfsdk:"all_users"`
	FilterID      types.String `tfsdk:"filter_id"`
	FilterType    types.String `tfsdk:"filter_type"`
}

// AssignmentModelAttrTypes returns the attribute types for AssignmentModel
func AssignmentModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"include_groups": types.ListType{ElemType: types.StringType},
		"exclude_groups": types.ListType{ElemType: types.StringType},
		"all_devices":    types.BoolType,
		"all_users":      types.BoolType,
		"filter_id":      types.StringType,
		"filter_type":    types.StringType,
	}
}

// AssignmentBlockSchema returns the schema for the assignment block
func AssignmentBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Assignment configuration for this policy. If not specified, the policy will not be assigned to any groups.",
		MarkdownDescription: `
Assignment configuration for this policy. Multiple assignment blocks can be specified.

Each assignment block can target:
- Specific Azure AD groups (include_groups)
- All devices (all_devices = true)
- All users (all_users = true)

Exclusions can be specified with exclude_groups.
`,
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
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
					Description: "Assign to all devices.",
					Optional:    true,
				},
				"all_users": schema.BoolAttribute{
					Description: "Assign to all licensed users.",
					Optional:    true,
				},
				"filter_id": schema.StringAttribute{
					Description: "The ID of an assignment filter to apply.",
					Optional:    true,
				},
				"filter_type": schema.StringAttribute{
					Description: "The type of filter: 'include' or 'exclude'.",
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.OneOf("include", "exclude"),
					},
				},
			},
		},
	}
}

// BuildAssignmentsFromBlocks builds assignment objects from assignment blocks
func BuildAssignmentsFromBlocks(ctx context.Context, assignments []AssignmentModel, diags *diag.Diagnostics) []clients.PolicyAssignment {
	var result []clients.PolicyAssignment

	for _, assignment := range assignments {
		// Handle include groups
		if !assignment.IncludeGroups.IsNull() {
			var groupIds []string
			diags.Append(assignment.IncludeGroups.ElementsAs(ctx, &groupIds, false)...)
			if diags.HasError() {
				return nil
			}

			for _, groupId := range groupIds {
				target := &clients.AssignmentTarget{
					ODataType: "#microsoft.graph.groupAssignmentTarget",
					GroupId:   groupId,
				}
				addFilterToTarget(target, assignment)
				result = append(result, clients.PolicyAssignment{Target: target})
			}
		}

		// Handle all devices
		if !assignment.AllDevices.IsNull() && assignment.AllDevices.ValueBool() {
			target := &clients.AssignmentTarget{
				ODataType: "#microsoft.graph.allDevicesAssignmentTarget",
			}
			addFilterToTarget(target, assignment)
			result = append(result, clients.PolicyAssignment{Target: target})
		}

		// Handle all users
		if !assignment.AllUsers.IsNull() && assignment.AllUsers.ValueBool() {
			target := &clients.AssignmentTarget{
				ODataType: "#microsoft.graph.allLicensedUsersAssignmentTarget",
			}
			addFilterToTarget(target, assignment)
			result = append(result, clients.PolicyAssignment{Target: target})
		}

		// Handle exclude groups
		if !assignment.ExcludeGroups.IsNull() {
			var groupIds []string
			diags.Append(assignment.ExcludeGroups.ElementsAs(ctx, &groupIds, false)...)
			if diags.HasError() {
				return nil
			}

			for _, groupId := range groupIds {
				result = append(result, clients.PolicyAssignment{
					Target: &clients.AssignmentTarget{
						ODataType: "#microsoft.graph.exclusionGroupAssignmentTarget",
						GroupId:   groupId,
					},
				})
			}
		}
	}

	return result
}

// addFilterToTarget adds filter configuration to an assignment target
func addFilterToTarget(target *clients.AssignmentTarget, assignment AssignmentModel) {
	if !assignment.FilterID.IsNull() && assignment.FilterID.ValueString() != "" {
		target.DeviceAndAppManagementAssignmentFilterId = assignment.FilterID.ValueString()
		filterType := "include"
		if !assignment.FilterType.IsNull() {
			filterType = assignment.FilterType.ValueString()
		}
		target.DeviceAndAppManagementAssignmentFilterType = filterType
	}
}

// AssignPolicy creates or updates policy assignments
func AssignPolicy(ctx context.Context, client *clients.GraphClient, policyType, policyId string, assignments []clients.PolicyAssignment) error {
	assignPath := getAssignPath(policyType, policyId)
	if assignPath == "" {
		return fmt.Errorf("unknown policy type: %s", policyType)
	}

	body := map[string]interface{}{
		"assignments": assignments,
	}

	_, err := client.Post(ctx, assignPath, body)
	if err != nil {
		return fmt.Errorf("failed to assign policy: %w", err)
	}

	tflog.Debug(ctx, "Assigned policy", map[string]interface{}{
		"policy_id":   policyId,
		"policy_type": policyType,
		"assignments": len(assignments),
	})

	return nil
}

// ReadPolicyAssignments reads the current assignments for a policy
func ReadPolicyAssignments(ctx context.Context, client *clients.GraphClient, policyType, policyId string) ([]AssignmentModel, error) {
	readPath := getAssignmentsReadPath(policyType, policyId)
	if readPath == "" {
		return nil, fmt.Errorf("unknown policy type: %s", policyType)
	}

	response, err := client.Get(ctx, readPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read assignments: %w", err)
	}

	if response.Value == nil {
		return nil, nil
	}

	var apiAssignments []struct {
		ID     string `json:"id"`
		Target struct {
			ODataType                                  string `json:"@odata.type"`
			GroupId                                    string `json:"groupId"`
			DeviceAndAppManagementAssignmentFilterId   string `json:"deviceAndAppManagementAssignmentFilterId"`
			DeviceAndAppManagementAssignmentFilterType string `json:"deviceAndAppManagementAssignmentFilterType"`
		} `json:"target"`
	}

	if err := json.Unmarshal(response.Value, &apiAssignments); err != nil {
		return nil, fmt.Errorf("failed to parse assignments: %w", err)
	}

	// Group assignments by type for building AssignmentModel objects
	var includeGroups []string
	var excludeGroups []string
	var allDevices, allUsers bool
	var filterID, filterType string

	for _, a := range apiAssignments {
		switch a.Target.ODataType {
		case "#microsoft.graph.groupAssignmentTarget":
			includeGroups = append(includeGroups, a.Target.GroupId)
			if a.Target.DeviceAndAppManagementAssignmentFilterId != "" {
				filterID = a.Target.DeviceAndAppManagementAssignmentFilterId
				filterType = a.Target.DeviceAndAppManagementAssignmentFilterType
			}
		case "#microsoft.graph.exclusionGroupAssignmentTarget":
			excludeGroups = append(excludeGroups, a.Target.GroupId)
		case "#microsoft.graph.allDevicesAssignmentTarget":
			allDevices = true
			if a.Target.DeviceAndAppManagementAssignmentFilterId != "" {
				filterID = a.Target.DeviceAndAppManagementAssignmentFilterId
				filterType = a.Target.DeviceAndAppManagementAssignmentFilterType
			}
		case "#microsoft.graph.allLicensedUsersAssignmentTarget":
			allUsers = true
			if a.Target.DeviceAndAppManagementAssignmentFilterId != "" {
				filterID = a.Target.DeviceAndAppManagementAssignmentFilterId
				filterType = a.Target.DeviceAndAppManagementAssignmentFilterType
			}
		}
	}

	// If no assignments, return nil
	if len(includeGroups) == 0 && len(excludeGroups) == 0 && !allDevices && !allUsers {
		return nil, nil
	}

	// Build a single AssignmentModel that represents all assignments
	assignment := AssignmentModel{
		AllDevices: types.BoolValue(allDevices),
		AllUsers:   types.BoolValue(allUsers),
	}

	if len(includeGroups) > 0 {
		includeList, _ := types.ListValueFrom(ctx, types.StringType, includeGroups)
		assignment.IncludeGroups = includeList
	} else {
		assignment.IncludeGroups = types.ListNull(types.StringType)
	}

	if len(excludeGroups) > 0 {
		excludeList, _ := types.ListValueFrom(ctx, types.StringType, excludeGroups)
		assignment.ExcludeGroups = excludeList
	} else {
		assignment.ExcludeGroups = types.ListNull(types.StringType)
	}

	if filterID != "" {
		assignment.FilterID = types.StringValue(filterID)
		assignment.FilterType = types.StringValue(filterType)
	} else {
		assignment.FilterID = types.StringNull()
		assignment.FilterType = types.StringNull()
	}

	return []AssignmentModel{assignment}, nil
}

// getAssignPath returns the API path for creating/updating assignments
func getAssignPath(policyType, policyId string) string {
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

// getAssignmentsReadPath returns the API path for reading assignments
func getAssignmentsReadPath(policyType, policyId string) string {
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
