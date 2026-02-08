// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/MANCHTOOLS/tofutune/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SettingsCatalogTemplateDataSource{}

// NewSettingsCatalogTemplateDataSource creates a new data source instance
func NewSettingsCatalogTemplateDataSource() datasource.DataSource {
	return &SettingsCatalogTemplateDataSource{}
}

// SettingsCatalogTemplateDataSource defines the data source implementation
type SettingsCatalogTemplateDataSource struct {
	client *clients.GraphClient
}

// SettingsCatalogTemplateDataSourceModel describes the data source data model
type SettingsCatalogTemplateDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	DisplayName     types.String `tfsdk:"display_name"`
	Description     types.String `tfsdk:"description"`
	BaseId          types.String `tfsdk:"base_id"`
	Version         types.Int64  `tfsdk:"version"`
	TemplateFamily  types.String `tfsdk:"template_family"`
	Platforms       types.String `tfsdk:"platforms"`
	Technologies    types.String `tfsdk:"technologies"`
	SettingCount    types.Int64  `tfsdk:"setting_count"`
}

// Metadata returns the data source type name
func (d *SettingsCatalogTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settings_catalog_template"
}

// Schema defines the schema for the data source
func (d *SettingsCatalogTemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a Settings Catalog template.",
		MarkdownDescription: `
Retrieves information about a Settings Catalog template.

Use this data source to look up template information for creating Settings Catalog policies
with predefined settings.

## Example Usage

` + "```hcl" + `
data "intune_settings_catalog_template" "windows_security" {
  display_name = "Windows Security"
}

resource "intune_settings_catalog_policy" "security" {
  name         = "Security Baseline"
  platforms    = "windows10AndLater"
  technologies = "mdm"
  template_id  = data.intune_settings_catalog_template.windows_security.id
}
` + "```" + `

## Template Families

Common template families include:
- ` + "`endpointSecurityAntivirus`" + ` - Microsoft Defender Antivirus settings
- ` + "`endpointSecurityDiskEncryption`" + ` - BitLocker settings
- ` + "`endpointSecurityFirewall`" + ` - Windows Firewall settings
- ` + "`endpointSecurityEndpointDetectionAndResponse`" + ` - EDR settings
- ` + "`endpointSecurityAttackSurfaceReduction`" + ` - ASR rules
- ` + "`endpointSecurityAccountProtection`" + ` - Windows Hello, Credential Guard
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the template.",
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name to search for. Performs a partial match.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the template.",
				Computed:    true,
			},
			"base_id": schema.StringAttribute{
				Description: "The base template ID.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "The template version.",
				Computed:    true,
			},
			"template_family": schema.StringAttribute{
				Description: "The template family (e.g., endpointSecurityAntivirus).",
				Computed:    true,
			},
			"platforms": schema.StringAttribute{
				Description: "The platforms this template supports.",
				Computed:    true,
			},
			"technologies": schema.StringAttribute{
				Description: "The technologies this template supports.",
				Computed:    true,
			},
			"setting_count": schema.Int64Attribute{
				Description: "The number of settings in this template.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *SettingsCatalogTemplateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SettingsCatalogTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SettingsCatalogTemplateDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	displayName := data.DisplayName.ValueString()

	tflog.Debug(ctx, "Reading settings catalog template", map[string]interface{}{
		"display_name": displayName,
	})

	// Get templates
	path := "/deviceManagement/configurationPolicyTemplates"
	items, err := d.client.ListAll(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Settings Catalog Templates",
			fmt.Sprintf("Could not list templates: %s", err),
		)
		return
	}

	// Find matching template
	var foundTemplate *struct {
		ID             string `json:"id"`
		DisplayName    string `json:"displayName"`
		Description    string `json:"description"`
		BaseId         string `json:"baseId"`
		Version        int    `json:"version"`
		TemplateFamily string `json:"templateFamily"`
		Platforms      string `json:"platforms"`
		Technologies   string `json:"technologies"`
		SettingCount   int    `json:"settingCount"`
	}

	for _, item := range items {
		var template struct {
			ID             string `json:"id"`
			DisplayName    string `json:"displayName"`
			Description    string `json:"description"`
			BaseId         string `json:"baseId"`
			Version        int    `json:"version"`
			TemplateFamily string `json:"templateFamily"`
			Platforms      string `json:"platforms"`
			Technologies   string `json:"technologies"`
			SettingCount   int    `json:"settingCount"`
		}
		if err := json.Unmarshal(item, &template); err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(template.DisplayName), strings.ToLower(displayName)) {
			foundTemplate = &template
			break
		}
	}

	if foundTemplate == nil {
		resp.Diagnostics.AddError(
			"Template Not Found",
			fmt.Sprintf("No template found matching '%s'", displayName),
		)
		return
	}

	// Update the model
	data.ID = types.StringValue(foundTemplate.ID)
	data.DisplayName = types.StringValue(foundTemplate.DisplayName)
	data.Description = types.StringValue(foundTemplate.Description)
	data.BaseId = types.StringValue(foundTemplate.BaseId)
	data.Version = types.Int64Value(int64(foundTemplate.Version))
	data.TemplateFamily = types.StringValue(foundTemplate.TemplateFamily)
	data.Platforms = types.StringValue(foundTemplate.Platforms)
	data.Technologies = types.StringValue(foundTemplate.Technologies)
	data.SettingCount = types.Int64Value(int64(foundTemplate.SettingCount))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
