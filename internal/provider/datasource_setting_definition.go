// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/MANCHTOOLS/tofutune/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SettingDefinitionDataSource{}

// NewSettingDefinitionDataSource creates a new data source instance
func NewSettingDefinitionDataSource() datasource.DataSource {
	return &SettingDefinitionDataSource{}
}

// SettingDefinitionDataSource defines the data source implementation
type SettingDefinitionDataSource struct {
	client *clients.GraphClient
}

// SettingDefinitionDataSourceModel describes the data source data model
type SettingDefinitionDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	DisplayName     types.String `tfsdk:"display_name"`
	Description     types.String `tfsdk:"description"`
	BaseUri         types.String `tfsdk:"base_uri"`
	OffsetUri       types.String `tfsdk:"offset_uri"`
	CategoryId      types.String `tfsdk:"category_id"`
	SettingUsage    types.String `tfsdk:"setting_usage"`
	Platform        types.String `tfsdk:"platform"`
	Technologies    types.String `tfsdk:"technologies"`
	Keywords        types.List   `tfsdk:"keywords"`
}

// Metadata returns the data source type name
func (d *SettingDefinitionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting_definition"
}

// Schema defines the schema for the data source
func (d *SettingDefinitionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a Settings Catalog setting definition.",
		MarkdownDescription: `
Retrieves information about a Settings Catalog setting definition.

Use this data source to look up the definition ID for a specific setting to use in
` + "`intune_settings_catalog_policy_settings`" + `.

## Example Usage

` + "```hcl" + `
data "intune_setting_definition" "defender_realtime" {
  name = "disablerealtimemonitoring"
}

resource "intune_settings_catalog_policy_settings" "defender" {
  policy_id = intune_settings_catalog_policy.example.id

  setting {
    definition_id = data.intune_setting_definition.defender_realtime.id
    value_type    = "boolean"
    value         = "false"
  }
}
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier (definition ID) for the setting. This is the value to use in setting blocks.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name to search for in setting definitions. This performs a partial match.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the setting.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the setting.",
				Computed:    true,
			},
			"base_uri": schema.StringAttribute{
				Description: "The base URI for the setting.",
				Computed:    true,
			},
			"offset_uri": schema.StringAttribute{
				Description: "The offset URI for the setting.",
				Computed:    true,
			},
			"category_id": schema.StringAttribute{
				Description: "The category ID for the setting.",
				Computed:    true,
			},
			"setting_usage": schema.StringAttribute{
				Description: "The setting usage type.",
				Computed:    true,
			},
			"platform": schema.StringAttribute{
				Description: "The platform this setting applies to.",
				Computed:    true,
			},
			"technologies": schema.StringAttribute{
				Description: "The technologies this setting applies to.",
				Computed:    true,
			},
			"keywords": schema.ListAttribute{
				Description: "Keywords associated with the setting.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *SettingDefinitionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read reads the data source
func (d *SettingDefinitionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SettingDefinitionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	tflog.Debug(ctx, "Reading setting definition", map[string]interface{}{
		"name": name,
	})

	// Search for the setting definition
	filter := fmt.Sprintf("contains(name,'%s')", name)
	definitions, err := d.client.ListSettingDefinitions(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Setting Definition",
			fmt.Sprintf("Could not search for setting definition: %s", err),
		)
		return
	}

	if len(definitions) == 0 {
		resp.Diagnostics.AddError(
			"Setting Definition Not Found",
			fmt.Sprintf("No setting definition found matching '%s'", name),
		)
		return
	}

	// Use the first match
	def := definitions[0]

	// Update the model
	data.ID = types.StringValue(def.ID)
	data.DisplayName = types.StringValue(def.DisplayName)
	data.Description = types.StringValue(def.Description)
	data.BaseUri = types.StringValue(def.BaseUri)
	data.OffsetUri = types.StringValue(def.OffsetUri)
	data.CategoryId = types.StringValue(def.CategoryId)
	data.SettingUsage = types.StringValue(def.SettingUsage)

	if def.Applicability != nil {
		data.Platform = types.StringValue(def.Applicability.Platform)
		data.Technologies = types.StringValue(def.Applicability.Technologies)
	}

	if len(def.Keywords) > 0 {
		keywords, diags := types.ListValueFrom(ctx, types.StringType, def.Keywords)
		resp.Diagnostics.Append(diags...)
		data.Keywords = keywords
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
