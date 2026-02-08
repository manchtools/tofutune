// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
var _ resource.Resource = &SettingsCatalogPolicySettingsResource{}
var _ resource.ResourceWithImportState = &SettingsCatalogPolicySettingsResource{}

// NewSettingsCatalogPolicySettingsResource creates a new resource instance
func NewSettingsCatalogPolicySettingsResource() resource.Resource {
	return &SettingsCatalogPolicySettingsResource{}
}

// SettingsCatalogPolicySettingsResource defines the resource implementation
type SettingsCatalogPolicySettingsResource struct {
	client *clients.GraphClient
}

// SettingsCatalogPolicySettingsResourceModel describes the resource data model
type SettingsCatalogPolicySettingsResourceModel struct {
	ID       types.String `tfsdk:"id"`
	PolicyID types.String `tfsdk:"policy_id"`
	Settings types.List   `tfsdk:"setting"`
}

// SettingModel represents a single setting in the policy
type SettingModel struct {
	DefinitionID types.String `tfsdk:"definition_id"`
	ValueType    types.String `tfsdk:"value_type"`
	Value        types.String `tfsdk:"value"`
	Children     types.List   `tfsdk:"children"`
}

// SettingModelAttrTypes returns the attribute types for SettingModel
func SettingModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"definition_id": types.StringType,
		"value_type":    types.StringType,
		"value":         types.StringType,
		"children":      types.ListType{ElemType: types.ObjectType{AttrTypes: ChildSettingModelAttrTypes()}},
	}
}

// ChildSettingModel represents a child setting (for choice or group settings)
type ChildSettingModel struct {
	DefinitionID types.String `tfsdk:"definition_id"`
	ValueType    types.String `tfsdk:"value_type"`
	Value        types.String `tfsdk:"value"`
}

// ChildSettingModelAttrTypes returns the attribute types for ChildSettingModel
func ChildSettingModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"definition_id": types.StringType,
		"value_type":    types.StringType,
		"value":         types.StringType,
	}
}

// Metadata returns the resource type name
func (r *SettingsCatalogPolicySettingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settings_catalog_policy_settings"
}

// Schema defines the schema for the resource
func (r *SettingsCatalogPolicySettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages settings within an Intune Settings Catalog policy. This resource allows modular " +
			"management of settings, enabling composition from multiple modules.",
		MarkdownDescription: `
Manages settings within an Intune Settings Catalog policy.

This resource allows modular management of settings within a policy. You can use multiple
instances of this resource (or modules that wrap it) to compose a complete policy from
separate, reusable components.

## Example Usage

### Basic Settings

` + "```hcl" + `
resource "intune_settings_catalog_policy" "example" {
  name         = "Example Policy"
  platforms    = "windows10AndLater"
  technologies = "mdm"
}

resource "intune_settings_catalog_policy_settings" "defender" {
  policy_id = intune_settings_catalog_policy.example.id

  setting {
    definition_id = "device_vendor_msft_defender_configuration_disablerealtimemonitoring"
    value_type    = "boolean"
    value         = "false"
  }

  setting {
    definition_id = "device_vendor_msft_defender_configuration_allowcloudprotection"
    value_type    = "choice"
    value         = "device_vendor_msft_defender_configuration_allowcloudprotection_1"
  }
}
` + "```" + `

### Using with Modules

` + "```hcl" + `
module "defender_settings" {
  source    = "./modules/defender"
  policy_id = intune_settings_catalog_policy.example.id

  real_time_protection = true
  cloud_protection     = "enabled"
}
` + "```" + `

## Value Types

- ` + "`string`" + `: A string value
- ` + "`integer`" + `: An integer value
- ` + "`boolean`" + `: A boolean value ("true" or "false")
- ` + "`choice`" + `: A choice from predefined options (use the choice value ID)
- ` + "`collection`" + `: A collection of values (JSON array as string)
- ` + "`group`" + `: A group of child settings
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this settings block (policy_id used as ID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "The ID of the Settings Catalog policy to add settings to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"setting": schema.ListNestedBlock{
				Description: "A setting to include in the policy.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"definition_id": schema.StringAttribute{
							Description: "The setting definition ID. This identifies the specific setting in the Settings Catalog.",
							Required:    true,
						},
						"value_type": schema.StringAttribute{
							Description: "The type of value. Valid values: string, integer, boolean, choice, collection, group.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("string", "integer", "boolean", "choice", "collection", "group"),
							},
						},
						"value": schema.StringAttribute{
							Description: "The value for the setting. For boolean, use 'true' or 'false'. " +
								"For choice, use the choice option ID. For collection, use a JSON array string.",
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"children": schema.ListNestedBlock{
							Description: "Child settings for choice or group settings.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"definition_id": schema.StringAttribute{
										Description: "The child setting definition ID.",
										Required:    true,
									},
									"value_type": schema.StringAttribute{
										Description: "The type of value for the child setting.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("string", "integer", "boolean", "choice"),
										},
									},
									"value": schema.StringAttribute{
										Description: "The value for the child setting.",
										Required:    true,
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
func (r *SettingsCatalogPolicySettingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// convertToAPISettings converts the Terraform model to API settings
func (r *SettingsCatalogPolicySettingsResource) convertToAPISettings(ctx context.Context, data *SettingsCatalogPolicySettingsResourceModel, diags *diag.Diagnostics) []clients.SettingsCatalogPolicySetting {
	var settings []SettingModel
	diags.Append(data.Settings.ElementsAs(ctx, &settings, false)...)
	if diags.HasError() {
		return nil
	}

	var apiSettings []clients.SettingsCatalogPolicySetting

	for _, setting := range settings {
		apiSetting := clients.SettingsCatalogPolicySetting{
			ODataType:       "#microsoft.graph.deviceManagementConfigurationSetting",
			SettingInstance: r.convertSettingInstance(ctx, setting, diags),
		}
		apiSettings = append(apiSettings, apiSetting)
	}

	return apiSettings
}

// convertSettingInstance converts a setting model to an API setting instance
func (r *SettingsCatalogPolicySettingsResource) convertSettingInstance(ctx context.Context, setting SettingModel, diags *diag.Diagnostics) *clients.SettingInstance {
	instance := &clients.SettingInstance{
		SettingDefinitionId: setting.DefinitionID.ValueString(),
	}

	valueType := setting.ValueType.ValueString()
	value := setting.Value.ValueString()

	switch valueType {
	case "string":
		instance.ODataType = "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance"
		instance.SimpleSettingValue = &clients.SimpleSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
			Value:     value,
		}

	case "integer":
		instance.ODataType = "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance"
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			diags.AddError("Invalid Integer Value", fmt.Sprintf("Could not parse '%s' as integer: %s", value, err))
			return nil
		}
		instance.SimpleSettingValue = &clients.SimpleSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationIntegerSettingValue",
			Value:     intVal,
		}

	case "boolean":
		instance.ODataType = "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance"
		boolVal := value == "true"
		instance.SimpleSettingValue = &clients.SimpleSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationBooleanSettingValue",
			Value:     boolVal,
		}

	case "choice":
		instance.ODataType = "#microsoft.graph.deviceManagementConfigurationChoiceSettingInstance"
		instance.ChoiceSettingValue = &clients.ChoiceSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationChoiceSettingValue",
			Value:     value,
		}

		// Handle children for choice settings
		if !setting.Children.IsNull() && len(setting.Children.Elements()) > 0 {
			var children []ChildSettingModel
			diags.Append(setting.Children.ElementsAs(ctx, &children, false)...)
			if diags.HasError() {
				return nil
			}

			for _, child := range children {
				childSetting := r.convertChildSetting(child, diags)
				if childSetting != nil {
					instance.ChoiceSettingValue.Children = append(instance.ChoiceSettingValue.Children, *childSetting)
				}
			}
		}

	case "collection":
		instance.ODataType = "#microsoft.graph.deviceManagementConfigurationSimpleSettingCollectionInstance"
		// Parse JSON array
		var values []interface{}
		if err := json.Unmarshal([]byte(value), &values); err != nil {
			diags.AddError("Invalid Collection Value", fmt.Sprintf("Could not parse '%s' as JSON array: %s", value, err))
			return nil
		}
		for _, v := range values {
			instance.SimpleSettingCollectionValue = append(instance.SimpleSettingCollectionValue, clients.SimpleSettingValue{
				ODataType: "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
				Value:     fmt.Sprintf("%v", v),
			})
		}

	case "group":
		instance.ODataType = "#microsoft.graph.deviceManagementConfigurationGroupSettingInstance"
		instance.GroupSettingValue = &clients.GroupSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationGroupSettingValue",
		}

		// Handle children for group settings
		if !setting.Children.IsNull() && len(setting.Children.Elements()) > 0 {
			var children []ChildSettingModel
			diags.Append(setting.Children.ElementsAs(ctx, &children, false)...)
			if diags.HasError() {
				return nil
			}

			for _, child := range children {
				childSetting := r.convertChildSetting(child, diags)
				if childSetting != nil {
					instance.GroupSettingValue.Children = append(instance.GroupSettingValue.Children, *childSetting)
				}
			}
		}
	}

	return instance
}

// convertChildSetting converts a child setting model to an API setting
func (r *SettingsCatalogPolicySettingsResource) convertChildSetting(child ChildSettingModel, diags *diag.Diagnostics) *clients.SettingsCatalogPolicySetting {
	childInstance := &clients.SettingInstance{
		SettingDefinitionId: child.DefinitionID.ValueString(),
	}

	valueType := child.ValueType.ValueString()
	value := child.Value.ValueString()

	switch valueType {
	case "string":
		childInstance.ODataType = "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance"
		childInstance.SimpleSettingValue = &clients.SimpleSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
			Value:     value,
		}

	case "integer":
		childInstance.ODataType = "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance"
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			diags.AddError("Invalid Integer Value", fmt.Sprintf("Could not parse '%s' as integer: %s", value, err))
			return nil
		}
		childInstance.SimpleSettingValue = &clients.SimpleSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationIntegerSettingValue",
			Value:     intVal,
		}

	case "boolean":
		childInstance.ODataType = "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance"
		boolVal := value == "true"
		childInstance.SimpleSettingValue = &clients.SimpleSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationBooleanSettingValue",
			Value:     boolVal,
		}

	case "choice":
		childInstance.ODataType = "#microsoft.graph.deviceManagementConfigurationChoiceSettingInstance"
		childInstance.ChoiceSettingValue = &clients.ChoiceSettingValue{
			ODataType: "#microsoft.graph.deviceManagementConfigurationChoiceSettingValue",
			Value:     value,
		}
	}

	return &clients.SettingsCatalogPolicySetting{
		ODataType:       "#microsoft.graph.deviceManagementConfigurationSetting",
		SettingInstance: childInstance,
	}
}

// convertAPISettingsToModel converts API settings back to the Terraform model format
func (r *SettingsCatalogPolicySettingsResource) convertAPISettingsToModel(ctx context.Context, apiSettings []clients.SettingsCatalogPolicySetting, diags *diag.Diagnostics) types.List {
	var settings []SettingModel

	for _, apiSetting := range apiSettings {
		if apiSetting.SettingInstance == nil {
			continue
		}

		instance := apiSetting.SettingInstance
		setting := SettingModel{
			DefinitionID: types.StringValue(instance.SettingDefinitionId),
			Children:     types.ListNull(types.ObjectType{AttrTypes: ChildSettingModelAttrTypes()}),
		}

		// Determine the value type and extract the value
		switch {
		case instance.SimpleSettingValue != nil:
			setting.ValueType, setting.Value = r.parseSimpleSettingValue(instance.SimpleSettingValue)

		case instance.ChoiceSettingValue != nil:
			setting.ValueType = types.StringValue("choice")
			setting.Value = types.StringValue(instance.ChoiceSettingValue.Value)
			// Handle children for choice settings
			if len(instance.ChoiceSettingValue.Children) > 0 {
				setting.Children = r.parseChildSettings(ctx, instance.ChoiceSettingValue.Children, diags)
			}

		case len(instance.SimpleSettingCollectionValue) > 0:
			setting.ValueType = types.StringValue("collection")
			var values []string
			for _, v := range instance.SimpleSettingCollectionValue {
				values = append(values, fmt.Sprintf("%v", v.Value))
			}
			jsonBytes, _ := json.Marshal(values)
			setting.Value = types.StringValue(string(jsonBytes))

		case instance.GroupSettingValue != nil:
			setting.ValueType = types.StringValue("group")
			setting.Value = types.StringNull()
			if len(instance.GroupSettingValue.Children) > 0 {
				setting.Children = r.parseChildSettings(ctx, instance.GroupSettingValue.Children, diags)
			}

		default:
			// Unknown setting type, skip
			tflog.Warn(ctx, "Unknown setting type", map[string]interface{}{
				"definition_id": instance.SettingDefinitionId,
				"odata_type":    instance.ODataType,
			})
			continue
		}

		settings = append(settings, setting)
	}

	// Convert to types.List
	if len(settings) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: SettingModelAttrTypes()})
	}

	settingsList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: SettingModelAttrTypes()}, settings)
	diags.Append(listDiags...)

	return settingsList
}

// parseSimpleSettingValue parses a simple setting value and returns the value type and value
func (r *SettingsCatalogPolicySettingsResource) parseSimpleSettingValue(ssv *clients.SimpleSettingValue) (types.String, types.String) {
	if ssv == nil {
		return types.StringNull(), types.StringNull()
	}

	switch ssv.ODataType {
	case "#microsoft.graph.deviceManagementConfigurationStringSettingValue":
		return types.StringValue("string"), types.StringValue(fmt.Sprintf("%v", ssv.Value))

	case "#microsoft.graph.deviceManagementConfigurationIntegerSettingValue":
		return types.StringValue("integer"), types.StringValue(fmt.Sprintf("%v", ssv.Value))

	case "#microsoft.graph.deviceManagementConfigurationBooleanSettingValue":
		boolVal, ok := ssv.Value.(bool)
		if !ok {
			return types.StringValue("boolean"), types.StringValue(fmt.Sprintf("%v", ssv.Value))
		}
		if boolVal {
			return types.StringValue("boolean"), types.StringValue("true")
		}
		return types.StringValue("boolean"), types.StringValue("false")

	default:
		// Treat unknown types as string
		return types.StringValue("string"), types.StringValue(fmt.Sprintf("%v", ssv.Value))
	}
}

// parseChildSettings parses child settings from the API format
func (r *SettingsCatalogPolicySettingsResource) parseChildSettings(ctx context.Context, apiChildren []clients.SettingsCatalogPolicySetting, diags *diag.Diagnostics) types.List {
	var children []ChildSettingModel

	for _, apiChild := range apiChildren {
		if apiChild.SettingInstance == nil {
			continue
		}

		instance := apiChild.SettingInstance
		child := ChildSettingModel{
			DefinitionID: types.StringValue(instance.SettingDefinitionId),
		}

		switch {
		case instance.SimpleSettingValue != nil:
			child.ValueType, child.Value = r.parseSimpleSettingValue(instance.SimpleSettingValue)

		case instance.ChoiceSettingValue != nil:
			child.ValueType = types.StringValue("choice")
			child.Value = types.StringValue(instance.ChoiceSettingValue.Value)

		default:
			continue
		}

		children = append(children, child)
	}

	if len(children) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: ChildSettingModelAttrTypes()})
	}

	childList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ChildSettingModelAttrTypes()}, children)
	diags.Append(listDiags...)

	return childList
}

// Create creates the resource and sets the initial Terraform state
func (r *SettingsCatalogPolicySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SettingsCatalogPolicySettingsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := data.PolicyID.ValueString()

	tflog.Debug(ctx, "Creating Settings Catalog policy settings", map[string]interface{}{
		"policy_id": policyID,
	})

	// Convert settings to API format
	apiSettings := r.convertToAPISettings(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the policy with the settings
	err := r.client.UpdateSettingsCatalogPolicySettings(ctx, policyID, apiSettings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Settings Catalog Policy Settings",
			fmt.Sprintf("Could not update policy settings: %s", err),
		)
		return
	}

	// Use policy ID as the resource ID
	data.ID = data.PolicyID

	tflog.Debug(ctx, "Created Settings Catalog policy settings", map[string]interface{}{
		"policy_id": policyID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *SettingsCatalogPolicySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SettingsCatalogPolicySettingsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := data.PolicyID.ValueString()

	tflog.Debug(ctx, "Reading Settings Catalog policy settings", map[string]interface{}{
		"policy_id": policyID,
	})

	// Get the policy with settings
	policy, err := r.client.GetSettingsCatalogPolicy(ctx, policyID)
	if err != nil {
		// Check if policy was deleted
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Settings Catalog Policy Settings",
			fmt.Sprintf("Could not read policy ID %s: %s", policyID, err),
		)
		return
	}

	// If no settings exist, remove the resource
	if len(policy.Settings) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert API settings back to Terraform model
	data.Settings = r.convertAPISettingsToModel(ctx, policy.Settings, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read Settings Catalog policy settings", map[string]interface{}{
		"policy_id":      policyID,
		"settings_count": len(policy.Settings),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state
func (r *SettingsCatalogPolicySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SettingsCatalogPolicySettingsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := data.PolicyID.ValueString()

	tflog.Debug(ctx, "Updating Settings Catalog policy settings", map[string]interface{}{
		"policy_id": policyID,
	})

	// Convert settings to API format
	apiSettings := r.convertToAPISettings(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the policy with the settings
	err := r.client.UpdateSettingsCatalogPolicySettings(ctx, policyID, apiSettings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Settings Catalog Policy Settings",
			fmt.Sprintf("Could not update policy settings: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *SettingsCatalogPolicySettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SettingsCatalogPolicySettingsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := data.PolicyID.ValueString()

	tflog.Debug(ctx, "Deleting Settings Catalog policy settings", map[string]interface{}{
		"policy_id": policyID,
	})

	// Clear settings by updating with an empty array
	err := r.client.UpdateSettingsCatalogPolicySettings(ctx, policyID, []clients.SettingsCatalogPolicySetting{})
	if err != nil {
		// Ignore not found errors during delete
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Settings Catalog Policy Settings",
			fmt.Sprintf("Could not clear policy settings: %s", err),
		)
		return
	}
}

// ImportState imports the resource state
func (r *SettingsCatalogPolicySettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import uses the policy ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), req.ID)...)
}
