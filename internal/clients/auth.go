// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// AuthConfig holds the authentication configuration matching azuread provider patterns
type AuthConfig struct {
	// Tenant ID (required for most auth methods)
	TenantID string

	// Client ID for Service Principal authentication
	ClientID string

	// Client Secret for Service Principal with secret authentication
	ClientSecret string

	// Client Certificate path for Service Principal with certificate
	ClientCertificatePath string

	// Client Certificate password (optional)
	ClientCertificatePassword string

	// Use Azure CLI for authentication
	UseAzureCLI bool

	// Use Managed Identity for authentication
	UseManagedIdentity bool

	// Managed Identity Client ID (for user-assigned managed identity)
	ManagedIdentityClientID string

	// OIDC Token for federated authentication
	OIDCToken string

	// OIDC Token File Path for federated authentication
	OIDCTokenFilePath string

	// OIDC Request URL for federated authentication (GitHub Actions)
	OIDCRequestURL string

	// OIDC Request Token for federated authentication (GitHub Actions)
	OIDCRequestToken string

	// Environment (public, usgovernment, china, germany)
	Environment string

	// Custom metadata host
	MetadataHost string

	// Auxiliary Tenant IDs for multi-tenant scenarios
	AuxiliaryTenantIDs []string
}

// AuthMethod represents the authentication method being used
type AuthMethod string

const (
	AuthMethodAzureCLI        AuthMethod = "azure_cli"
	AuthMethodManagedIdentity AuthMethod = "managed_identity"
	AuthMethodClientSecret    AuthMethod = "client_secret"
	AuthMethodClientCert      AuthMethod = "client_certificate"
	AuthMethodOIDC            AuthMethod = "oidc"
)

// Authenticator provides Azure authentication credentials
type Authenticator struct {
	credential azcore.TokenCredential
	config     *AuthConfig
	method     AuthMethod
}

// NewAuthenticator creates a new authenticator based on the provided configuration
func NewAuthenticator(ctx context.Context, config *AuthConfig) (*Authenticator, error) {
	if config == nil {
		return nil, errors.New("authentication configuration is required")
	}

	auth := &Authenticator{
		config: config,
	}

	// Try authentication methods in order of precedence (matching azuread provider)
	// 1. OIDC (for CI/CD pipelines)
	// 2. Client Certificate
	// 3. Client Secret
	// 4. Managed Identity
	// 5. Azure CLI

	var err error

	// Check for OIDC authentication
	if config.OIDCToken != "" || config.OIDCTokenFilePath != "" || (config.OIDCRequestURL != "" && config.OIDCRequestToken != "") {
		auth.credential, err = auth.createOIDCCredential(ctx)
		if err == nil {
			auth.method = AuthMethodOIDC
			return auth, nil
		}
	}

	// Check for Client Certificate authentication
	if config.ClientCertificatePath != "" && config.ClientID != "" && config.TenantID != "" {
		auth.credential, err = auth.createClientCertificateCredential()
		if err == nil {
			auth.method = AuthMethodClientCert
			return auth, nil
		}
	}

	// Check for Client Secret authentication
	if config.ClientSecret != "" && config.ClientID != "" && config.TenantID != "" {
		auth.credential, err = auth.createClientSecretCredential()
		if err == nil {
			auth.method = AuthMethodClientSecret
			return auth, nil
		}
	}

	// Check for Managed Identity authentication
	if config.UseManagedIdentity {
		auth.credential, err = auth.createManagedIdentityCredential()
		if err == nil {
			auth.method = AuthMethodManagedIdentity
			return auth, nil
		}
	}

	// Fall back to Azure CLI authentication
	if config.UseAzureCLI || auth.credential == nil {
		auth.credential, err = auth.createAzureCLICredential()
		if err == nil {
			auth.method = AuthMethodAzureCLI
			return auth, nil
		}
	}

	// If we still don't have a credential, try default credential chain
	if auth.credential == nil {
		auth.credential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create any authentication credential: %w", err)
		}
		auth.method = AuthMethodAzureCLI // Default falls back to CLI-like behavior
	}

	return auth, nil
}

// GetToken retrieves an access token for the specified scopes
func (a *Authenticator) GetToken(ctx context.Context, scopes []string) (string, error) {
	token, err := a.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: scopes,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	return token.Token, nil
}

// GetCredential returns the underlying Azure credential
func (a *Authenticator) GetCredential() azcore.TokenCredential {
	return a.credential
}

// GetMethod returns the authentication method being used
func (a *Authenticator) GetMethod() AuthMethod {
	return a.method
}

// createOIDCCredential creates an OIDC credential for federated authentication
func (a *Authenticator) createOIDCCredential(ctx context.Context) (azcore.TokenCredential, error) {
	var token string

	// Get token from various sources
	if a.config.OIDCToken != "" {
		token = a.config.OIDCToken
	} else if a.config.OIDCTokenFilePath != "" {
		tokenBytes, err := os.ReadFile(a.config.OIDCTokenFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read OIDC token file: %w", err)
		}
		token = strings.TrimSpace(string(tokenBytes))
	} else if a.config.OIDCRequestURL != "" && a.config.OIDCRequestToken != "" {
		// GitHub Actions OIDC - fetch token from GitHub's OIDC provider
		var err error
		token, err = a.fetchGitHubOIDCToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GitHub OIDC token: %w", err)
		}
	}

	if token == "" {
		return nil, errors.New("no OIDC token available")
	}

	// Create the client assertion credential
	cred, err := azidentity.NewClientAssertionCredential(
		a.config.TenantID,
		a.config.ClientID,
		func(ctx context.Context) (string, error) {
			return token, nil
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC credential: %w", err)
	}

	return cred, nil
}

// fetchGitHubOIDCToken fetches an OIDC token from GitHub Actions
func (a *Authenticator) fetchGitHubOIDCToken(ctx context.Context) (string, error) {
	// This would make an HTTP request to GitHub's OIDC provider
	// For now, we'll return an error indicating this needs implementation
	// In a real implementation, you would:
	// 1. Make a GET request to a.config.OIDCRequestURL
	// 2. Include Authorization header with "bearer " + a.config.OIDCRequestToken
	// 3. Parse the JSON response to get the "value" field containing the token

	return "", errors.New("GitHub OIDC token fetch not yet implemented - provide token directly via oidc_token")
}

// createClientSecretCredential creates a client secret credential
func (a *Authenticator) createClientSecretCredential() (azcore.TokenCredential, error) {
	cred, err := azidentity.NewClientSecretCredential(
		a.config.TenantID,
		a.config.ClientID,
		a.config.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client secret credential: %w", err)
	}
	return cred, nil
}

// createClientCertificateCredential creates a client certificate credential
func (a *Authenticator) createClientCertificateCredential() (azcore.TokenCredential, error) {
	certData, err := os.ReadFile(a.config.ClientCertificatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse the certificate
	var certs []*x509.Certificate
	var key interface{}

	// Try to parse as PEM
	for {
		block, rest := pem.Decode(certData)
		if block == nil {
			break
		}
		certData = rest

		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}
			certs = append(certs, cert)
		case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
			if a.config.ClientCertificatePassword != "" {
				key, err = x509.DecryptPEMBlock(block, []byte(a.config.ClientCertificatePassword))
			} else {
				key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
					if err != nil {
						key, err = x509.ParseECPrivateKey(block.Bytes)
					}
				}
			}
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}
		}
	}

	if len(certs) == 0 || key == nil {
		return nil, errors.New("certificate file must contain both certificate and private key")
	}

	cred, err := azidentity.NewClientCertificateCredential(
		a.config.TenantID,
		a.config.ClientID,
		certs,
		key,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client certificate credential: %w", err)
	}
	return cred, nil
}

// createManagedIdentityCredential creates a managed identity credential
func (a *Authenticator) createManagedIdentityCredential() (azcore.TokenCredential, error) {
	opts := &azidentity.ManagedIdentityCredentialOptions{}

	// If a specific client ID is provided, use user-assigned managed identity
	if a.config.ManagedIdentityClientID != "" {
		opts.ID = azidentity.ClientID(a.config.ManagedIdentityClientID)
	}

	cred, err := azidentity.NewManagedIdentityCredential(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed identity credential: %w", err)
	}
	return cred, nil
}

// createAzureCLICredential creates an Azure CLI credential
func (a *Authenticator) createAzureCLICredential() (azcore.TokenCredential, error) {
	opts := &azidentity.AzureCLICredentialOptions{}

	if a.config.TenantID != "" {
		opts.TenantID = a.config.TenantID
	}

	cred, err := azidentity.NewAzureCLICredential(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure CLI credential: %w", err)
	}
	return cred, nil
}
