// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/tofutune/tofutune/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &PolicyDataSource{}

// NewPolicyDataSource creates a new data source instance
func NewPolicyDataSource() datasource.DataSource {
	return &PolicyDataSource{}
}

// PolicyDataSource defines the data source implementation
type PolicyDataSource struct {
	client *clients.GraphClient
}

// PolicyDataSourceModel describes the data source data model
type PolicyDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	DisplayName          types.String `tfsdk:"display_name"`
	Description          types.String `tfsdk:"description"`
	PolicyType           types.String `tfsdk:"policy_type"`
	Platforms            types.String `tfsdk:"platforms"`
	Technologies         types.String `tfsdk:"technologies"`
	CreatedDateTime      types.String `tfsdk:"created_date_time"`
	LastModifiedDateTime types.String `tfsdk:"last_modified_date_time"`
	RoleScopeTagIds      types.List   `tfsdk:"role_scope_tag_ids"`
}

// Metadata returns the data source type name
func (d *PolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

// Schema defines the schema for the data source
func (d *PolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing Intune policy.",
		MarkdownDescription: `
Retrieves information about an existing Intune policy.

Use this data source to reference policies that were created outside of Terraform/OpenTofu,
or to get information about policies managed elsewhere.

## Example Usage

### By Display Name

` + "```hcl" + `
data "intune_policy" "existing" {
  display_name = "Corporate Security Baseline"
  policy_type  = "settings_catalog"
}

# Use the policy ID for assignments
resource "intune_policy_assignment" "assign" {
  policy_id   = data.intune_policy.existing.id
  policy_type = "settings_catalog"
  all_devices = true
}
` + "```" + `

### By Policy ID

` + "```hcl" + `
data "intune_policy" "by_id" {
  id          = "00000000-0000-0000-0000-000000000000"
  policy_type = "compliance"
}
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the policy. Either id or display_name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the policy. Either id or display_name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the policy.",
				Computed:    true,
			},
			"policy_type": schema.StringAttribute{
				Description: "The type of policy to search for. Valid values: settings_catalog, compliance, endpoint_security, device_configuration.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						PolicyTypeSettingsCatalog,
						PolicyTypeCompliance,
						PolicyTypeEndpointSecurity,
						PolicyTypeDeviceConfig,
					),
				},
			},
			"platforms": schema.StringAttribute{
				Description: "The platforms the policy applies to.",
				Computed:    true,
			},
			"technologies": schema.StringAttribute{
				Description: "The technologies the policy uses.",
				Computed:    true,
			},
			"created_date_time": schema.StringAttribute{
				Description: "The date and time the policy was created.",
				Computed:    true,
			},
			"last_modified_date_time": schema.StringAttribute{
				Description: "The date and time the policy was last modified.",
				Computed:    true,
			},
			"role_scope_tag_ids": schema.ListAttribute{
				Description: "The role scope tag IDs assigned to the policy.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *PolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *PolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyType := data.PolicyType.ValueString()
	id := data.ID.ValueString()
	displayName := data.DisplayName.ValueString()

	if id == "" && displayName == "" {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either id or display_name must be specified.",
		)
		return
	}

	tflog.Debug(ctx, "Reading policy", map[string]interface{}{
		"policy_type":  policyType,
		"id":           id,
		"display_name": displayName,
	})

	// Determine the API path based on policy type
	var basePath string
	switch policyType {
	case PolicyTypeSettingsCatalog:
		basePath = "/deviceManagement/configurationPolicies"
	case PolicyTypeCompliance:
		basePath = "/deviceManagement/deviceCompliancePolicies"
	case PolicyTypeEndpointSecurity:
		basePath = "/deviceManagement/intents"
	case PolicyTypeDeviceConfig:
		basePath = "/deviceManagement/deviceConfigurations"
	default:
		resp.Diagnostics.AddError(
			"Invalid Policy Type",
			fmt.Sprintf("Unknown policy type: %s", policyType),
		)
		return
	}

	var policyData map[string]interface{}

	if id != "" {
		// Get policy by ID
		path := basePath
		if policyType == PolicyTypeSettingsCatalog {
			path = fmt.Sprintf("%s('%s')", basePath, id)
		} else {
			path = fmt.Sprintf("%s/%s", basePath, id)
		}

		response, err := d.client.Get(ctx, path)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Policy",
				fmt.Sprintf("Could not read policy: %s", err),
			)
			return
		}

		respBytes, _ := json.Marshal(response)
		if err := json.Unmarshal(respBytes, &policyData); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Response",
				fmt.Sprintf("Could not parse policy response: %s", err),
			)
			return
		}

		if policyData["id"] == nil || policyData["id"] == "" {
			policyData["id"] = response.ID
		}
	} else {
		// Search by display name
		items, err := d.client.ListAll(ctx, basePath)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Policies",
				fmt.Sprintf("Could not list policies: %s", err),
			)
			return
		}

		for _, item := range items {
			var policy map[string]interface{}
			if err := json.Unmarshal(item, &policy); err != nil {
				continue
			}

			// Check display name (different field name for different policy types)
			var name string
			if n, ok := policy["displayName"].(string); ok {
				name = n
			} else if n, ok := policy["name"].(string); ok {
				name = n
			}

			if strings.EqualFold(name, displayName) {
				policyData = policy
				break
			}
		}

		if policyData == nil {
			resp.Diagnostics.AddError(
				"Policy Not Found",
				fmt.Sprintf("No policy found with display name '%s'", displayName),
			)
			return
		}
	}

	// Update the model from the response
	if id, ok := policyData["id"].(string); ok {
		data.ID = types.StringValue(id)
	}

	// Handle different field names for different policy types
	if name, ok := policyData["displayName"].(string); ok {
		data.DisplayName = types.StringValue(name)
	} else if name, ok := policyData["name"].(string); ok {
		data.DisplayName = types.StringValue(name)
	}

	if desc, ok := policyData["description"].(string); ok {
		data.Description = types.StringValue(desc)
	}

	if platforms, ok := policyData["platforms"].(string); ok {
		data.Platforms = types.StringValue(platforms)
	}

	if technologies, ok := policyData["technologies"].(string); ok {
		data.Technologies = types.StringValue(technologies)
	}

	if created, ok := policyData["createdDateTime"].(string); ok {
		data.CreatedDateTime = types.StringValue(created)
	}

	if modified, ok := policyData["lastModifiedDateTime"].(string); ok {
		data.LastModifiedDateTime = types.StringValue(modified)
	}

	if tagIds, ok := policyData["roleScopeTagIds"].([]interface{}); ok {
		var tags []string
		for _, t := range tagIds {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
		if len(tags) > 0 {
			tagList, diags := types.ListValueFrom(ctx, types.StringType, tags)
			resp.Diagnostics.Append(diags...)
			data.RoleScopeTagIds = tagList
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
