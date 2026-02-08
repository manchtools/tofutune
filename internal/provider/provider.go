// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/tofutune/tofutune/internal/clients"
)

// Ensure IntuneProvider satisfies various provider interfaces
var _ provider.Provider = &IntuneProvider{}

// IntuneProvider defines the provider implementation
type IntuneProvider struct {
	version string
}

// IntuneProviderModel describes the provider data model
type IntuneProviderModel struct {
	// Tenant configuration
	TenantID    types.String `tfsdk:"tenant_id"`
	Environment types.String `tfsdk:"environment"`

	// Service Principal authentication
	ClientID                  types.String `tfsdk:"client_id"`
	ClientSecret              types.String `tfsdk:"client_secret"`
	ClientCertificatePath     types.String `tfsdk:"client_certificate_path"`
	ClientCertificatePassword types.String `tfsdk:"client_certificate_password"`

	// Managed Identity authentication
	UseManagedIdentity      types.Bool   `tfsdk:"use_msi"`
	ManagedIdentityClientID types.String `tfsdk:"msi_client_id"`

	// Azure CLI authentication
	UseAzureCLI types.Bool `tfsdk:"use_cli"`

	// OIDC authentication
	UseOIDC           types.Bool   `tfsdk:"use_oidc"`
	OIDCToken         types.String `tfsdk:"oidc_token"`
	OIDCTokenFilePath types.String `tfsdk:"oidc_token_file_path"`
	OIDCRequestURL    types.String `tfsdk:"oidc_request_url"`
	OIDCRequestToken  types.String `tfsdk:"oidc_request_token"`

	// Multi-tenant
	AuxiliaryTenantIDs types.List `tfsdk:"auxiliary_tenant_ids"`

	// Metadata
	MetadataHost types.String `tfsdk:"metadata_host"`
}

// ProviderData contains the configured clients for resources
type ProviderData struct {
	GraphClient *clients.GraphClient
	Auth        *clients.Authenticator
}

// New creates a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &IntuneProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name
func (p *IntuneProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "intune"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *IntuneProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Intune provider allows you to manage Microsoft Intune resources using OpenTofu/Terraform. " +
			"Authentication is compatible with the hashicorp/azuread provider.",
		MarkdownDescription: `
The Intune provider allows you to manage Microsoft Intune resources using OpenTofu/Terraform.

## Authentication

This provider supports the same authentication methods as the hashicorp/azuread provider:

* **Azure CLI** - Uses credentials from ` + "`az login`" + `
* **Service Principal with Client Secret** - Uses a client ID and secret
* **Service Principal with Client Certificate** - Uses a client ID and certificate
* **Managed Identity** - For Azure-hosted resources
* **OIDC** - For GitHub Actions and similar CI/CD systems

## Required Permissions

The following Microsoft Graph API permissions are required:

* ` + "`DeviceManagementConfiguration.ReadWrite.All`" + ` - For Settings Catalog policies
* ` + "`DeviceManagementManagedDevices.ReadWrite.All`" + ` - For compliance policies
* ` + "`Group.Read.All`" + ` - For policy assignments
`,
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				Description: "The Tenant ID which should be used. This can also be sourced from the ARM_TENANT_ID environment variable.",
				Optional:    true,
			},
			"environment": schema.StringAttribute{
				Description: "The cloud environment to use. Possible values are: public, usgovernment, china. Defaults to public. " +
					"This can also be sourced from the ARM_ENVIRONMENT environment variable.",
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				Description: "The Client ID which should be used for service principal authentication. " +
					"This can also be sourced from the ARM_CLIENT_ID environment variable.",
				Optional: true,
			},
			"client_secret": schema.StringAttribute{
				Description: "The Client Secret which should be used for service principal authentication. " +
					"This can also be sourced from the ARM_CLIENT_SECRET environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"client_certificate_path": schema.StringAttribute{
				Description: "The path to the Client Certificate (PEM format) for service principal authentication. " +
					"This can also be sourced from the ARM_CLIENT_CERTIFICATE_PATH environment variable.",
				Optional: true,
			},
			"client_certificate_password": schema.StringAttribute{
				Description: "The password for the Client Certificate. " +
					"This can also be sourced from the ARM_CLIENT_CERTIFICATE_PASSWORD environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"use_msi": schema.BoolAttribute{
				Description: "Should Managed Identity be used for authentication? " +
					"This can also be sourced from the ARM_USE_MSI environment variable. Defaults to false.",
				Optional: true,
			},
			"msi_client_id": schema.StringAttribute{
				Description: "The Client ID of the User Assigned Managed Identity. " +
					"This can also be sourced from the ARM_MSI_CLIENT_ID environment variable.",
				Optional: true,
			},
			"use_cli": schema.BoolAttribute{
				Description: "Should Azure CLI be used for authentication? " +
					"This can also be sourced from the ARM_USE_CLI environment variable. Defaults to true.",
				Optional: true,
			},
			"use_oidc": schema.BoolAttribute{
				Description: "Should OIDC be used for authentication? " +
					"This can also be sourced from the ARM_USE_OIDC environment variable. Defaults to false.",
				Optional: true,
			},
			"oidc_token": schema.StringAttribute{
				Description: "The OIDC token to use for authentication. " +
					"This can also be sourced from the ARM_OIDC_TOKEN environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"oidc_token_file_path": schema.StringAttribute{
				Description: "The path to a file containing an OIDC token. " +
					"This can also be sourced from the ARM_OIDC_TOKEN_FILE_PATH environment variable.",
				Optional: true,
			},
			"oidc_request_url": schema.StringAttribute{
				Description: "The URL for the OIDC provider (GitHub Actions). " +
					"This can also be sourced from the ACTIONS_ID_TOKEN_REQUEST_URL environment variable.",
				Optional: true,
			},
			"oidc_request_token": schema.StringAttribute{
				Description: "The bearer token for the OIDC provider (GitHub Actions). " +
					"This can also be sourced from the ACTIONS_ID_TOKEN_REQUEST_TOKEN environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"auxiliary_tenant_ids": schema.ListAttribute{
				Description: "A list of additional Tenant IDs for multi-tenant authentication.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"metadata_host": schema.StringAttribute{
				Description: "The hostname which should be used for the Azure Metadata Service.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares the provider for data sources and resources
func (p *IntuneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Intune provider")

	var config IntuneProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build authentication configuration from provider config and environment variables
	authConfig := &clients.AuthConfig{}

	// Tenant ID
	if !config.TenantID.IsNull() {
		authConfig.TenantID = config.TenantID.ValueString()
	} else if v := os.Getenv("ARM_TENANT_ID"); v != "" {
		authConfig.TenantID = v
	}

	// Environment
	if !config.Environment.IsNull() {
		authConfig.Environment = config.Environment.ValueString()
	} else if v := os.Getenv("ARM_ENVIRONMENT"); v != "" {
		authConfig.Environment = v
	}

	// Client ID
	if !config.ClientID.IsNull() {
		authConfig.ClientID = config.ClientID.ValueString()
	} else if v := os.Getenv("ARM_CLIENT_ID"); v != "" {
		authConfig.ClientID = v
	}

	// Client Secret
	if !config.ClientSecret.IsNull() {
		authConfig.ClientSecret = config.ClientSecret.ValueString()
	} else if v := os.Getenv("ARM_CLIENT_SECRET"); v != "" {
		authConfig.ClientSecret = v
	}

	// Client Certificate
	if !config.ClientCertificatePath.IsNull() {
		authConfig.ClientCertificatePath = config.ClientCertificatePath.ValueString()
	} else if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PATH"); v != "" {
		authConfig.ClientCertificatePath = v
	}

	if !config.ClientCertificatePassword.IsNull() {
		authConfig.ClientCertificatePassword = config.ClientCertificatePassword.ValueString()
	} else if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PASSWORD"); v != "" {
		authConfig.ClientCertificatePassword = v
	}

	// Managed Identity
	if !config.UseManagedIdentity.IsNull() {
		authConfig.UseManagedIdentity = config.UseManagedIdentity.ValueBool()
	} else if v := os.Getenv("ARM_USE_MSI"); v == "true" {
		authConfig.UseManagedIdentity = true
	}

	if !config.ManagedIdentityClientID.IsNull() {
		authConfig.ManagedIdentityClientID = config.ManagedIdentityClientID.ValueString()
	} else if v := os.Getenv("ARM_MSI_CLIENT_ID"); v != "" {
		authConfig.ManagedIdentityClientID = v
	}

	// Azure CLI
	if !config.UseAzureCLI.IsNull() {
		authConfig.UseAzureCLI = config.UseAzureCLI.ValueBool()
	} else if v := os.Getenv("ARM_USE_CLI"); v == "true" {
		authConfig.UseAzureCLI = true
	} else {
		// Default to true if no other auth method is specified
		authConfig.UseAzureCLI = true
	}

	// OIDC
	useOIDC := false
	if !config.UseOIDC.IsNull() {
		useOIDC = config.UseOIDC.ValueBool()
	} else if v := os.Getenv("ARM_USE_OIDC"); v == "true" {
		useOIDC = true
	}

	if useOIDC {
		if !config.OIDCToken.IsNull() {
			authConfig.OIDCToken = config.OIDCToken.ValueString()
		} else if v := os.Getenv("ARM_OIDC_TOKEN"); v != "" {
			authConfig.OIDCToken = v
		}

		if !config.OIDCTokenFilePath.IsNull() {
			authConfig.OIDCTokenFilePath = config.OIDCTokenFilePath.ValueString()
		} else if v := os.Getenv("ARM_OIDC_TOKEN_FILE_PATH"); v != "" {
			authConfig.OIDCTokenFilePath = v
		}

		if !config.OIDCRequestURL.IsNull() {
			authConfig.OIDCRequestURL = config.OIDCRequestURL.ValueString()
		} else if v := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL"); v != "" {
			authConfig.OIDCRequestURL = v
		}

		if !config.OIDCRequestToken.IsNull() {
			authConfig.OIDCRequestToken = config.OIDCRequestToken.ValueString()
		} else if v := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"); v != "" {
			authConfig.OIDCRequestToken = v
		}
	}

	// Metadata Host
	if !config.MetadataHost.IsNull() {
		authConfig.MetadataHost = config.MetadataHost.ValueString()
	}

	// Auxiliary Tenant IDs
	if !config.AuxiliaryTenantIDs.IsNull() {
		var tenantIDs []string
		resp.Diagnostics.Append(config.AuxiliaryTenantIDs.ElementsAs(ctx, &tenantIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		authConfig.AuxiliaryTenantIDs = tenantIDs
	}

	// Validate required configuration
	if authConfig.TenantID == "" && !authConfig.UseAzureCLI {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant_id"),
			"Missing Tenant ID",
			"The provider cannot create the Intune API client as there is a missing or empty value for the tenant ID. "+
				"Set the tenant_id value in the configuration or use the ARM_TENANT_ID environment variable. "+
				"If using Azure CLI authentication, the tenant can be inferred from the CLI context.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create authenticator
	auth, err := clients.NewAuthenticator(ctx, authConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Authenticator",
			fmt.Sprintf("An unexpected error occurred when creating the authenticator: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Authentication configured", map[string]interface{}{
		"method": string(auth.GetMethod()),
	})

	// Create Graph client
	userAgent := fmt.Sprintf("TofuTune/%s", p.version)
	graphClient := clients.NewGraphClient(auth, userAgent)

	// Create provider data
	providerData := &ProviderData{
		GraphClient: graphClient,
		Auth:        auth,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData

	tflog.Info(ctx, "Intune provider configured successfully")
}

// Resources defines the resources implemented in the provider
func (p *IntuneProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSettingsCatalogPolicyResource,
		NewSettingsCatalogPolicySettingsResource,
		NewCompliancePolicyResource,
		NewEndpointSecurityPolicyResource,
		NewPolicyAssignmentResource,
	}
}

// DataSources defines the data sources implemented in the provider
func (p *IntuneProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSettingDefinitionDataSource,
		NewSettingsCatalogTemplateDataSource,
		NewPolicyDataSource,
	}
}
