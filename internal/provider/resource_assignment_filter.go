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

	"github.com/MANCHTOOLS/tofutune/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &AssignmentFilterResource{}
var _ resource.ResourceWithImportState = &AssignmentFilterResource{}

// NewAssignmentFilterResource returns a new assignment filter resource
func NewAssignmentFilterResource() resource.Resource {
	return &AssignmentFilterResource{}
}

// AssignmentFilterResource defines the resource implementation
type AssignmentFilterResource struct {
	client *clients.GraphClient
}

// AssignmentFilterResourceModel describes the resource data model
type AssignmentFilterResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	DisplayName          types.String `tfsdk:"display_name"`
	Description          types.String `tfsdk:"description"`
	Platform             types.String `tfsdk:"platform"`
	Rule                 types.String `tfsdk:"rule"`
	RoleScopeTags        types.List   `tfsdk:"role_scope_tags"`
	CreatedDateTime      types.String `tfsdk:"created_date_time"`
	LastModifiedDateTime types.String `tfsdk:"last_modified_date_time"`
}

// Metadata returns the resource type name
func (r *AssignmentFilterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assignment_filter"
}

// Schema defines the schema for the resource
func (r *AssignmentFilterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Intune assignment filter.",
		MarkdownDescription: `
Manages an Intune assignment filter.

Assignment filters allow you to dynamically include or exclude devices from policy and app assignments
based on device properties. Filters use a rule syntax to evaluate device properties at assignment time.

## Example Usage

### Filter by Device Model

` + "```hcl" + `
resource "intune_assignment_filter" "surface_devices" {
  display_name = "Surface Devices"
  description  = "Filter for Microsoft Surface devices"
  platform     = "windows10AndLater"
  rule         = "(device.model -startsWith \"Surface\")"
}
` + "```" + `

### Filter by Manufacturer

` + "```hcl" + `
resource "intune_assignment_filter" "dell_devices" {
  display_name = "Dell Devices"
  description  = "Filter for Dell manufactured devices"
  platform     = "windows10AndLater"
  rule         = "(device.manufacturer -eq \"Dell Inc.\")"
}
` + "```" + `

### Filter by OS Version

` + "```hcl" + `
resource "intune_assignment_filter" "windows_11" {
  display_name = "Windows 11 Devices"
  description  = "Filter for Windows 11 devices"
  platform     = "windows10AndLater"
  rule         = "(device.osVersion -startsWith \"10.0.22\")"
}
` + "```" + `

### Complex Filter with Multiple Conditions

` + "```hcl" + `
resource "intune_assignment_filter" "corporate_laptops" {
  display_name = "Corporate Laptops"
  description  = "Filter for corporate-owned laptop devices"
  platform     = "windows10AndLater"
  rule         = "(device.deviceOwnership -eq \"Corporate\") and (device.deviceCategory -eq \"Laptop\")"
}
` + "```" + `

### Use Filter in Policy Assignment

` + "```hcl" + `
resource "intune_assignment_filter" "surface_pro" {
  display_name = "Surface Pro Devices"
  platform     = "windows10AndLater"
  rule         = "(device.model -contains \"Surface Pro\")"
}

resource "intune_settings_catalog_policy" "surface_settings" {
  name         = "Surface Pro Settings"
  platforms    = "windows10AndLater"
  technologies = "mdm"

  assignment {
    all_devices = true
    filter_id   = intune_assignment_filter.surface_pro.id
    filter_type = "include"
  }
}
` + "```" + `

## Filter Rule Syntax

Filter rules use a syntax similar to Azure AD dynamic group rules. Common operators include:

| Operator | Description |
|----------|-------------|
| ` + "`-eq`" + ` | Equals |
| ` + "`-ne`" + ` | Not equals |
| ` + "`-startsWith`" + ` | Starts with |
| ` + "`-contains`" + ` | Contains |
| ` + "`-in`" + ` | In array |
| ` + "`-notIn`" + ` | Not in array |

Common device properties for filtering:

- ` + "`device.deviceOwnership`" + ` - Corporate, Personal
- ` + "`device.manufacturer`" + ` - Device manufacturer
- ` + "`device.model`" + ` - Device model
- ` + "`device.osVersion`" + ` - Operating system version
- ` + "`device.deviceCategory`" + ` - Device category
- ` + "`device.enrollmentProfileName`" + ` - Enrollment profile name
- ` + "`device.operatingSystemSKU`" + ` - OS SKU

## Import

Assignment filters can be imported using the filter ID:

` + "```shell" + `
terraform import intune_assignment_filter.example 00000000-0000-0000-0000-000000000000
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the assignment filter.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the assignment filter.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the assignment filter.",
				Optional:    true,
			},
			"platform": schema.StringAttribute{
				Description: "The platform for which this filter applies.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"android",
						"androidForWork",
						"iOS",
						"macOS",
						"windows10AndLater",
						"androidWorkProfile",
						"androidAOSP",
						"androidMobileApplicationManagement",
						"iOSMobileApplicationManagement",
						"windowsMobileApplicationManagement",
						"unknown",
					),
				},
			},
			"rule": schema.StringAttribute{
				Description: "The rule expression that determines which devices match this filter.",
				Required:    true,
			},
			"role_scope_tags": schema.ListAttribute{
				Description: "The list of role scope tag IDs for this filter.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"created_date_time": schema.StringAttribute{
				Description: "The date and time the filter was created.",
				Computed:    true,
			},
			"last_modified_date_time": schema.StringAttribute{
				Description: "The date and time the filter was last modified.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *AssignmentFilterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AssignmentFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AssignmentFilterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	filter := &clients.AssignmentFilter{
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
		Platform:    data.Platform.ValueString(),
		Rule:        data.Rule.ValueString(),
	}

	// Handle role scope tags
	if !data.RoleScopeTags.IsNull() {
		var tags []string
		resp.Diagnostics.Append(data.RoleScopeTags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		filter.RoleScopeTags = tags
	}

	// Create the assignment filter
	created, err := r.client.CreateAssignmentFilter(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Assignment Filter",
			fmt.Sprintf("Could not create assignment filter: %s", err),
		)
		return
	}

	// Update the model with the created filter data
	data.ID = types.StringValue(created.ID)
	data.DisplayName = types.StringValue(created.DisplayName)
	data.Description = types.StringValue(created.Description)
	data.Platform = types.StringValue(created.Platform)
	data.Rule = types.StringValue(created.Rule)
	data.CreatedDateTime = types.StringValue(created.CreatedDateTime)
	data.LastModifiedDateTime = types.StringValue(created.LastModifiedDateTime)

	// Handle role scope tags
	if len(created.RoleScopeTags) > 0 {
		tagList, diags := types.ListValueFrom(ctx, types.StringType, created.RoleScopeTags)
		resp.Diagnostics.Append(diags...)
		data.RoleScopeTags = tagList
	} else {
		data.RoleScopeTags = types.ListNull(types.StringType)
	}

	tflog.Debug(ctx, "Created assignment filter", map[string]interface{}{
		"id":           created.ID,
		"display_name": created.DisplayName,
		"platform":     created.Platform,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *AssignmentFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AssignmentFilterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter, err := r.client.GetAssignmentFilter(ctx, data.ID.ValueString())
	if err != nil {
		// Check if the resource was deleted outside of Terraform
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Assignment Filter",
			fmt.Sprintf("Could not read assignment filter ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update the model with the read data
	data.DisplayName = types.StringValue(filter.DisplayName)
	data.Description = types.StringValue(filter.Description)
	data.Platform = types.StringValue(filter.Platform)
	data.Rule = types.StringValue(filter.Rule)
	data.CreatedDateTime = types.StringValue(filter.CreatedDateTime)
	data.LastModifiedDateTime = types.StringValue(filter.LastModifiedDateTime)

	// Handle role scope tags
	if len(filter.RoleScopeTags) > 0 {
		tagList, diags := types.ListValueFrom(ctx, types.StringType, filter.RoleScopeTags)
		resp.Diagnostics.Append(diags...)
		data.RoleScopeTags = tagList
	} else {
		data.RoleScopeTags = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state
func (r *AssignmentFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AssignmentFilterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	filter := &clients.AssignmentFilter{
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
		Platform:    data.Platform.ValueString(),
		Rule:        data.Rule.ValueString(),
	}

	// Handle role scope tags
	if !data.RoleScopeTags.IsNull() {
		var tags []string
		resp.Diagnostics.Append(data.RoleScopeTags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		filter.RoleScopeTags = tags
	}

	// Update the assignment filter
	updated, err := r.client.UpdateAssignmentFilter(ctx, data.ID.ValueString(), filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Assignment Filter",
			fmt.Sprintf("Could not update assignment filter ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update the model with the updated filter data
	data.DisplayName = types.StringValue(updated.DisplayName)
	data.Description = types.StringValue(updated.Description)
	data.Platform = types.StringValue(updated.Platform)
	data.Rule = types.StringValue(updated.Rule)
	data.LastModifiedDateTime = types.StringValue(updated.LastModifiedDateTime)

	// Handle role scope tags
	if len(updated.RoleScopeTags) > 0 {
		tagList, diags := types.ListValueFrom(ctx, types.StringType, updated.RoleScopeTags)
		resp.Diagnostics.Append(diags...)
		data.RoleScopeTags = tagList
	} else {
		data.RoleScopeTags = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *AssignmentFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AssignmentFilterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAssignmentFilter(ctx, data.ID.ValueString())
	if err != nil {
		// Ignore "not found" errors as the resource is already deleted
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Assignment Filter",
			fmt.Sprintf("Could not delete assignment filter ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

// ImportState imports the resource state
func (r *AssignmentFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
