// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/MANCHTOOLS/tofutune/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &AssignmentFiltersDataSource{}

// NewAssignmentFiltersDataSource returns a new assignment filters data source
func NewAssignmentFiltersDataSource() datasource.DataSource {
	return &AssignmentFiltersDataSource{}
}

// AssignmentFiltersDataSource defines the data source implementation
type AssignmentFiltersDataSource struct {
	client *clients.GraphClient
}

// AssignmentFilterDataModel describes a single assignment filter
type AssignmentFilterDataModel struct {
	ID                   types.String `tfsdk:"id"`
	DisplayName          types.String `tfsdk:"display_name"`
	Description          types.String `tfsdk:"description"`
	Platform             types.String `tfsdk:"platform"`
	Rule                 types.String `tfsdk:"rule"`
	CreatedDateTime      types.String `tfsdk:"created_date_time"`
	LastModifiedDateTime types.String `tfsdk:"last_modified_date_time"`
}

// AssignmentFiltersDataSourceModel describes the data source data model
type AssignmentFiltersDataSourceModel struct {
	Platform types.String                `tfsdk:"platform"`
	Filters  []AssignmentFilterDataModel `tfsdk:"filters"`
}

// Metadata returns the data source type name
func (d *AssignmentFiltersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assignment_filters"
}

// Schema defines the schema for the data source
func (d *AssignmentFiltersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves Intune assignment filters.",
		MarkdownDescription: `
Retrieves Intune assignment filters.

This data source returns a list of assignment filters, optionally filtered by platform.
Assignment filters can be referenced in policy assignments to dynamically include or exclude devices.

## Example Usage

### List All Assignment Filters

` + "```hcl" + `
data "intune_assignment_filters" "all" {}

output "all_filters" {
  value = data.intune_assignment_filters.all.filters
}
` + "```" + `

### Filter by Platform

` + "```hcl" + `
data "intune_assignment_filters" "windows" {
  platform = "windows10AndLater"
}

output "windows_filters" {
  value = data.intune_assignment_filters.windows.filters
}
` + "```" + `

### Use Existing Filter in Policy

` + "```hcl" + `
data "intune_assignment_filters" "all" {}

locals {
  surface_filter = [
    for filter in data.intune_assignment_filters.all.filters : filter
    if filter.display_name == "Surface Devices"
  ][0]
}

resource "intune_settings_catalog_policy" "surface_policy" {
  name         = "Surface Device Settings"
  platforms    = "windows10AndLater"
  technologies = "mdm"

  assignment {
    all_devices = true
    filter_id   = local.surface_filter.id
    filter_type = "include"
  }
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"platform": schema.StringAttribute{
				Description: "Optional platform filter. If specified, only filters for this platform are returned.",
				Optional:    true,
			},
			"filters": schema.ListNestedAttribute{
				Description: "List of assignment filters.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the assignment filter.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the assignment filter.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the assignment filter.",
							Computed:    true,
						},
						"platform": schema.StringAttribute{
							Description: "The platform for the assignment filter.",
							Computed:    true,
						},
						"rule": schema.StringAttribute{
							Description: "The rule expression for the assignment filter.",
							Computed:    true,
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
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *AssignmentFiltersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.GraphClient
}

// Read refreshes the Terraform state with the latest data
func (d *AssignmentFiltersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AssignmentFiltersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve all assignment filters
	filters, err := d.client.ListAssignmentFilters(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Assignment Filters",
			fmt.Sprintf("Could not read assignment filters: %s", err),
		)
		return
	}

	// Filter by platform if specified
	platformFilter := data.Platform.ValueString()

	// Map the API response to the data model
	var filteredResults []AssignmentFilterDataModel
	for _, filter := range filters {
		// Skip if platform filter is set and doesn't match
		if platformFilter != "" && filter.Platform != platformFilter {
			continue
		}

		filteredResults = append(filteredResults, AssignmentFilterDataModel{
			ID:                   types.StringValue(filter.ID),
			DisplayName:          types.StringValue(filter.DisplayName),
			Description:          types.StringValue(filter.Description),
			Platform:             types.StringValue(filter.Platform),
			Rule:                 types.StringValue(filter.Rule),
			CreatedDateTime:      types.StringValue(filter.CreatedDateTime),
			LastModifiedDateTime: types.StringValue(filter.LastModifiedDateTime),
		})
	}

	data.Filters = filteredResults

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
