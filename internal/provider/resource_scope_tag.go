// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/MANCHTOOLS/tofutune/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ScopeTagResource{}
var _ resource.ResourceWithImportState = &ScopeTagResource{}

// NewScopeTagResource returns a new scope tag resource
func NewScopeTagResource() resource.Resource {
	return &ScopeTagResource{}
}

// ScopeTagResource defines the resource implementation
type ScopeTagResource struct {
	client *clients.GraphClient
}

// ScopeTagResourceModel describes the resource data model
type ScopeTagResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	IsBuiltIn   types.Bool   `tfsdk:"is_built_in"`
}

// Metadata returns the resource type name
func (r *ScopeTagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scope_tag"
}

// Schema defines the schema for the resource
func (r *ScopeTagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Intune role scope tag.",
		MarkdownDescription: `
Manages an Intune role scope tag.

Role scope tags allow you to limit the visibility of Intune objects to specific administrators.
Tags are used to control which objects administrators can see and manage in Intune.

## Example Usage

### Basic Scope Tag

` + "```hcl" + `
resource "intune_scope_tag" "engineering" {
  display_name = "Engineering"
  description  = "Scope tag for engineering team devices and policies"
}
` + "```" + `

### Use Scope Tag in Policies

` + "```hcl" + `
resource "intune_scope_tag" "sales" {
  display_name = "Sales Department"
  description  = "Scope tag for sales department devices"
}

resource "intune_settings_catalog_policy" "sales_policy" {
  name             = "Sales Device Configuration"
  description      = "Settings for sales devices"
  platforms        = "windows10AndLater"
  technologies     = "mdm"
  role_scope_tag_ids = [intune_scope_tag.sales.id]
}
` + "```" + `

## Import

Scope tags can be imported using the scope tag ID:

` + "```shell" + `
terraform import intune_scope_tag.example 00000000-0000-0000-0000-000000000000
` + "```" + `

~> **Note:** The default scope tag (ID "0") is built-in and cannot be managed by Terraform.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the scope tag.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the scope tag.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the scope tag.",
				Optional:    true,
			},
			"is_built_in": schema.BoolAttribute{
				Description: "Indicates whether this scope tag is built-in (default scope tag).",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ScopeTagResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ScopeTagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ScopeTagResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	tag := &clients.ScopeTag{
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
	}

	// Create the scope tag
	created, err := r.client.CreateScopeTag(ctx, tag)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Scope Tag",
			fmt.Sprintf("Could not create scope tag: %s", err),
		)
		return
	}

	// Update the model with the created scope tag data
	data.ID = types.StringValue(created.ID)
	data.DisplayName = types.StringValue(created.DisplayName)
	data.Description = types.StringValue(created.Description)
	data.IsBuiltIn = types.BoolValue(created.IsBuiltIn)

	tflog.Debug(ctx, "Created scope tag", map[string]interface{}{
		"id":           created.ID,
		"display_name": created.DisplayName,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *ScopeTagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ScopeTagResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.client.GetScopeTag(ctx, data.ID.ValueString())
	if err != nil {
		// Check if the resource was deleted outside of Terraform
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Scope Tag",
			fmt.Sprintf("Could not read scope tag ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update the model with the read data
	data.DisplayName = types.StringValue(tag.DisplayName)
	data.Description = types.StringValue(tag.Description)
	data.IsBuiltIn = types.BoolValue(tag.IsBuiltIn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state
func (r *ScopeTagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ScopeTagResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	tag := &clients.ScopeTag{
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
	}

	// Update the scope tag
	updated, err := r.client.UpdateScopeTag(ctx, data.ID.ValueString(), tag)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Scope Tag",
			fmt.Sprintf("Could not update scope tag ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Update the model with the updated scope tag data
	data.DisplayName = types.StringValue(updated.DisplayName)
	data.Description = types.StringValue(updated.Description)
	data.IsBuiltIn = types.BoolValue(updated.IsBuiltIn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *ScopeTagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScopeTagResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if this is the built-in scope tag
	if data.ID.ValueString() == "0" {
		resp.Diagnostics.AddError(
			"Cannot Delete Built-in Scope Tag",
			"The default scope tag (ID 0) is built-in and cannot be deleted.",
		)
		return
	}

	err := r.client.DeleteScopeTag(ctx, data.ID.ValueString())
	if err != nil {
		// Ignore "not found" errors as the resource is already deleted
		if graphErr, ok := err.(*clients.GraphError); ok && graphErr.Code == "NotFound" {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Scope Tag",
			fmt.Sprintf("Could not delete scope tag ID %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

// ImportState imports the resource state
func (r *ScopeTagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
