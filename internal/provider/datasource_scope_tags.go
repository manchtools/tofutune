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
var _ datasource.DataSource = &ScopeTagsDataSource{}

// NewScopeTagsDataSource returns a new scope tags data source
func NewScopeTagsDataSource() datasource.DataSource {
	return &ScopeTagsDataSource{}
}

// ScopeTagsDataSource defines the data source implementation
type ScopeTagsDataSource struct {
	client *clients.GraphClient
}

// ScopeTagDataModel describes a single scope tag
type ScopeTagDataModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	IsBuiltIn   types.Bool   `tfsdk:"is_built_in"`
}

// ScopeTagsDataSourceModel describes the data source data model
type ScopeTagsDataSourceModel struct {
	ScopeTags []ScopeTagDataModel `tfsdk:"scope_tags"`
}

// Metadata returns the data source type name
func (d *ScopeTagsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scope_tags"
}

// Schema defines the schema for the data source
func (d *ScopeTagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves all Intune role scope tags.",
		MarkdownDescription: `
Retrieves all Intune role scope tags.

This data source returns a list of all scope tags in the Intune tenant, which can be used
to reference existing scope tags in policy configurations.

## Example Usage

### List All Scope Tags

` + "```hcl" + `
data "intune_scope_tags" "all" {}

output "all_scope_tags" {
  value = data.intune_scope_tags.all.scope_tags
}
` + "```" + `

### Find a Specific Scope Tag

` + "```hcl" + `
data "intune_scope_tags" "all" {}

locals {
  engineering_tag = [
    for tag in data.intune_scope_tags.all.scope_tags : tag
    if tag.display_name == "Engineering"
  ][0]
}

resource "intune_settings_catalog_policy" "engineering_policy" {
  name             = "Engineering Device Configuration"
  platforms        = "windows10AndLater"
  technologies     = "mdm"
  role_scope_tag_ids = [local.engineering_tag.id]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"scope_tags": schema.ListNestedAttribute{
				Description: "List of scope tags.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the scope tag.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the scope tag.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the scope tag.",
							Computed:    true,
						},
						"is_built_in": schema.BoolAttribute{
							Description: "Indicates whether this scope tag is built-in.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *ScopeTagsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ScopeTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ScopeTagsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve all scope tags
	tags, err := d.client.ListScopeTags(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Scope Tags",
			fmt.Sprintf("Could not read scope tags: %s", err),
		)
		return
	}

	// Map the API response to the data model
	data.ScopeTags = make([]ScopeTagDataModel, len(tags))
	for i, tag := range tags {
		data.ScopeTags[i] = ScopeTagDataModel{
			ID:          types.StringValue(tag.ID),
			DisplayName: types.StringValue(tag.DisplayName),
			Description: types.StringValue(tag.Description),
			IsBuiltIn:   types.BoolValue(tag.IsBuiltIn),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
