// Copyright (c) TofuTune Contributors
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultGraphEndpoint is the Microsoft Graph API endpoint
	DefaultGraphEndpoint = "https://graph.microsoft.com"

	// GraphAPIVersion is the API version to use (beta needed for many Intune features)
	GraphAPIVersion = "beta"

	// GraphScope is the scope required for Microsoft Graph API
	GraphScope = "https://graph.microsoft.com/.default"
)

// GraphClient provides access to Microsoft Graph API for Intune operations
type GraphClient struct {
	auth       *Authenticator
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// NewGraphClient creates a new Graph API client
func NewGraphClient(auth *Authenticator, userAgent string) *GraphClient {
	return &GraphClient{
		auth:       auth,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		baseURL:    fmt.Sprintf("%s/%s", DefaultGraphEndpoint, GraphAPIVersion),
		userAgent:  userAgent,
	}
}

// GraphResponse represents a generic Graph API response
type GraphResponse struct {
	ODataContext  string          `json:"@odata.context,omitempty"`
	ODataType     string          `json:"@odata.type,omitempty"`
	ODataNextLink string          `json:"@odata.nextLink,omitempty"`
	Value         json.RawMessage `json:"value,omitempty"`
	ID            string          `json:"id,omitempty"`
	Error         *GraphError     `json:"error,omitempty"`
}

// GraphError represents an error from the Graph API
type GraphError struct {
	Code       string       `json:"code"`
	Message    string       `json:"message"`
	InnerError *GraphError  `json:"innerError,omitempty"`
	Details    []GraphError `json:"details,omitempty"`
}

func (e *GraphError) Error() string {
	if e.InnerError != nil {
		return fmt.Sprintf("%s: %s (inner: %s)", e.Code, e.Message, e.InnerError.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// doRequest performs an HTTP request to the Graph API
func (c *GraphClient) doRequest(ctx context.Context, method, path string, body interface{}) (*GraphResponse, error) {
	// Get access token
	token, err := c.auth.GetToken(ctx, []string{GraphScope})
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Build URL
	reqURL := fmt.Sprintf("%s%s", c.baseURL, path)

	// Prepare body
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var graphResp GraphResponse
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &graphResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
		}
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		if graphResp.Error != nil {
			return nil, graphResp.Error
		}
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return &graphResp, nil
}

// Get performs a GET request
func (c *GraphClient) Get(ctx context.Context, path string) (*GraphResponse, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *GraphClient) Post(ctx context.Context, path string, body interface{}) (*GraphResponse, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

// Patch performs a PATCH request
func (c *GraphClient) Patch(ctx context.Context, path string, body interface{}) (*GraphResponse, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body)
}

// Put performs a PUT request
func (c *GraphClient) Put(ctx context.Context, path string, body interface{}) (*GraphResponse, error) {
	return c.doRequest(ctx, http.MethodPut, path, body)
}

// Delete performs a DELETE request
func (c *GraphClient) Delete(ctx context.Context, path string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// ListAll retrieves all items from a paginated endpoint
func (c *GraphClient) ListAll(ctx context.Context, path string) ([]json.RawMessage, error) {
	var allItems []json.RawMessage
	currentPath := path

	for {
		resp, err := c.Get(ctx, currentPath)
		if err != nil {
			return nil, err
		}

		// Parse the value array
		var items []json.RawMessage
		if resp.Value != nil {
			if err := json.Unmarshal(resp.Value, &items); err != nil {
				return nil, fmt.Errorf("failed to parse items: %w", err)
			}
			allItems = append(allItems, items...)
		}

		// Check for next page
		if resp.ODataNextLink == "" {
			break
		}

		// Extract path from next link
		nextURL, err := url.Parse(resp.ODataNextLink)
		if err != nil {
			return nil, fmt.Errorf("failed to parse next link: %w", err)
		}
		currentPath = nextURL.Path + "?" + nextURL.RawQuery
		// Remove the version prefix if present
		currentPath = strings.TrimPrefix(currentPath, "/"+GraphAPIVersion)
	}

	return allItems, nil
}

// SettingsCatalogPolicy represents an Intune Settings Catalog policy
type SettingsCatalogPolicy struct {
	ODataType            string                           `json:"@odata.type,omitempty"`
	ID                   string                           `json:"id,omitempty"`
	Name                 string                           `json:"name"`
	Description          string                           `json:"description,omitempty"`
	Platforms            string                           `json:"platforms"`
	Technologies         string                           `json:"technologies"`
	CreatedDateTime      string                           `json:"createdDateTime,omitempty"`
	LastModifiedDateTime string                           `json:"lastModifiedDateTime,omitempty"`
	RoleScopeTagIds      []string                         `json:"roleScopeTagIds,omitempty"`
	SettingCount         int                              `json:"settingCount,omitempty"`
	Settings             []SettingsCatalogPolicySetting   `json:"settings,omitempty"`
	TemplateReference    *SettingsCatalogTemplateReference `json:"templateReference,omitempty"`
}

// SettingsCatalogPolicySetting represents a setting within a Settings Catalog policy
type SettingsCatalogPolicySetting struct {
	ODataType          string                  `json:"@odata.type,omitempty"`
	ID                 string                  `json:"id,omitempty"`
	SettingInstance    *SettingInstance        `json:"settingInstance"`
}

// SettingInstance represents a setting instance configuration
type SettingInstance struct {
	ODataType                  string                     `json:"@odata.type"`
	SettingDefinitionId        string                     `json:"settingDefinitionId"`
	SettingInstanceTemplateRef *SettingInstanceTemplateRef `json:"settingInstanceTemplateReference,omitempty"`
	// For simple value settings
	SimpleSettingValue         *SimpleSettingValue         `json:"simpleSettingValue,omitempty"`
	// For choice settings
	ChoiceSettingValue         *ChoiceSettingValue         `json:"choiceSettingValue,omitempty"`
	// For collection settings
	SimpleSettingCollectionValue []SimpleSettingValue      `json:"simpleSettingCollectionValue,omitempty"`
	// For group settings
	GroupSettingValue          *GroupSettingValue          `json:"groupSettingValue,omitempty"`
	GroupSettingCollectionValue []GroupSettingValue        `json:"groupSettingCollectionValue,omitempty"`
}

// SettingInstanceTemplateRef references a setting instance template
type SettingInstanceTemplateRef struct {
	SettingInstanceTemplateId string `json:"settingInstanceTemplateId,omitempty"`
}

// SimpleSettingValue represents a simple setting value
type SimpleSettingValue struct {
	ODataType string      `json:"@odata.type"`
	Value     interface{} `json:"value"`
}

// ChoiceSettingValue represents a choice setting value
type ChoiceSettingValue struct {
	ODataType string                           `json:"@odata.type,omitempty"`
	Value     string                           `json:"value"`
	Children  []SettingsCatalogPolicySetting   `json:"children,omitempty"`
}

// GroupSettingValue represents a group setting value
type GroupSettingValue struct {
	ODataType string                         `json:"@odata.type,omitempty"`
	Children  []SettingsCatalogPolicySetting `json:"children,omitempty"`
}

// SettingsCatalogTemplateReference references a settings catalog template
type SettingsCatalogTemplateReference struct {
	TemplateId       string `json:"templateId,omitempty"`
	TemplateFamily   string `json:"templateFamily,omitempty"`
	TemplateDisplayName string `json:"templateDisplayName,omitempty"`
	TemplateDisplayVersion string `json:"templateDisplayVersion,omitempty"`
}

// CompliancePolicy represents an Intune device compliance policy
type CompliancePolicy struct {
	ODataType                       string                      `json:"@odata.type,omitempty"`
	ID                              string                      `json:"id,omitempty"`
	DisplayName                     string                      `json:"displayName"`
	Description                     string                      `json:"description,omitempty"`
	CreatedDateTime                 string                      `json:"createdDateTime,omitempty"`
	LastModifiedDateTime            string                      `json:"lastModifiedDateTime,omitempty"`
	RoleScopeTagIds                 []string                    `json:"roleScopeTagIds,omitempty"`
	Version                         int                         `json:"version,omitempty"`
	ScheduledActionsForRule         []ComplianceScheduledAction `json:"scheduledActionsForRule,omitempty"`
	// Windows 10 specific settings
	PasswordRequired                bool   `json:"passwordRequired,omitempty"`
	PasswordBlockSimple             bool   `json:"passwordBlockSimple,omitempty"`
	PasswordRequiredToUnlockFromIdle bool  `json:"passwordRequiredToUnlockFromIdle,omitempty"`
	PasswordMinutesOfInactivityBeforeLock *int `json:"passwordMinutesOfInactivityBeforeLock,omitempty"`
	PasswordExpirationDays          *int   `json:"passwordExpirationDays,omitempty"`
	PasswordMinimumLength           *int   `json:"passwordMinimumLength,omitempty"`
	PasswordMinimumCharacterSetCount *int  `json:"passwordMinimumCharacterSetCount,omitempty"`
	PasswordRequiredType            string `json:"passwordRequiredType,omitempty"`
	PasswordPreviousPasswordBlockCount *int `json:"passwordPreviousPasswordBlockCount,omitempty"`
	RequireHealthyDeviceReport      bool   `json:"requireHealthyDeviceReport,omitempty"`
	OsMinimumVersion                string `json:"osMinimumVersion,omitempty"`
	OsMaximumVersion                string `json:"osMaximumVersion,omitempty"`
	MobileOsMinimumVersion          string `json:"mobileOsMinimumVersion,omitempty"`
	MobileOsMaximumVersion          string `json:"mobileOsMaximumVersion,omitempty"`
	EarlyLaunchAntiMalwareDriverEnabled bool `json:"earlyLaunchAntiMalwareDriverEnabled,omitempty"`
	BitLockerEnabled                bool   `json:"bitLockerEnabled,omitempty"`
	SecureBootEnabled               bool   `json:"secureBootEnabled,omitempty"`
	CodeIntegrityEnabled            bool   `json:"codeIntegrityEnabled,omitempty"`
	StorageRequireEncryption        bool   `json:"storageRequireEncryption,omitempty"`
	ActiveFirewallRequired          bool   `json:"activeFirewallRequired,omitempty"`
	DefenderEnabled                 bool   `json:"defenderEnabled,omitempty"`
	DefenderVersion                 string `json:"defenderVersion,omitempty"`
	SignatureOutOfDate              bool   `json:"signatureOutOfDate,omitempty"`
	RtpEnabled                      bool   `json:"rtpEnabled,omitempty"`
	AntivirusRequired               bool   `json:"antivirusRequired,omitempty"`
	AntiSpywareRequired             bool   `json:"antiSpywareRequired,omitempty"`
	DeviceThreatProtectionEnabled   bool   `json:"deviceThreatProtectionEnabled,omitempty"`
	DeviceThreatProtectionRequiredSecurityLevel string `json:"deviceThreatProtectionRequiredSecurityLevel,omitempty"`
	ConfigurationManagerComplianceRequired bool `json:"configurationManagerComplianceRequired,omitempty"`
	TpmRequired                     bool   `json:"tpmRequired,omitempty"`
	DeviceCompliancePolicyScript    *DeviceCompliancePolicyScript `json:"deviceCompliancePolicyScript,omitempty"`
	ValidOperatingSystemBuildRanges []OperatingSystemVersionRange `json:"validOperatingSystemBuildRanges,omitempty"`
}

// ComplianceScheduledAction represents a scheduled action for compliance
type ComplianceScheduledAction struct {
	RuleName                      string                       `json:"ruleName,omitempty"`
	ScheduledActionConfigurations []ScheduledActionConfiguration `json:"scheduledActionConfigurations,omitempty"`
}

// ScheduledActionConfiguration represents a scheduled action configuration
type ScheduledActionConfiguration struct {
	ID                       string   `json:"id,omitempty"`
	ActionType               string   `json:"actionType"`
	GracePeriodHours         int      `json:"gracePeriodHours"`
	NotificationTemplateId   string   `json:"notificationTemplateId,omitempty"`
	NotificationMessageCCList []string `json:"notificationMessageCCList,omitempty"`
}

// DeviceCompliancePolicyScript represents a compliance script
type DeviceCompliancePolicyScript struct {
	DeviceComplianceScriptId string `json:"deviceComplianceScriptId,omitempty"`
	RulesContent             string `json:"rulesContent,omitempty"`
}

// OperatingSystemVersionRange represents an OS version range
type OperatingSystemVersionRange struct {
	Description         string `json:"description,omitempty"`
	LowestVersion       string `json:"lowestVersion,omitempty"`
	HighestVersion      string `json:"highestVersion,omitempty"`
}

// EndpointSecurityPolicy represents an endpoint security policy
type EndpointSecurityPolicy struct {
	ODataType            string   `json:"@odata.type,omitempty"`
	ID                   string   `json:"id,omitempty"`
	DisplayName          string   `json:"displayName"`
	Description          string   `json:"description,omitempty"`
	Platforms            string   `json:"platforms"`
	Technologies         string   `json:"technologies"`
	TemplateId           string   `json:"templateId,omitempty"`
	CreatedDateTime      string   `json:"createdDateTime,omitempty"`
	LastModifiedDateTime string   `json:"lastModifiedDateTime,omitempty"`
	RoleScopeTagIds      []string `json:"roleScopeTagIds,omitempty"`
	IsAssigned           bool     `json:"isAssigned,omitempty"`
}

// PolicyAssignment represents a policy assignment
type PolicyAssignment struct {
	ODataType string           `json:"@odata.type,omitempty"`
	ID        string           `json:"id,omitempty"`
	Target    *AssignmentTarget `json:"target"`
	Source    string           `json:"source,omitempty"`
	SourceId  string           `json:"sourceId,omitempty"`
}

// AssignmentTarget represents an assignment target
type AssignmentTarget struct {
	ODataType                              string `json:"@odata.type"`
	DeviceAndAppManagementAssignmentFilterId   string `json:"deviceAndAppManagementAssignmentFilterId,omitempty"`
	DeviceAndAppManagementAssignmentFilterType string `json:"deviceAndAppManagementAssignmentFilterType,omitempty"`
	GroupId                                string `json:"groupId,omitempty"`
}

// SettingDefinition represents a setting definition from the Settings Catalog
type SettingDefinition struct {
	ODataType            string   `json:"@odata.type,omitempty"`
	ID                   string   `json:"id,omitempty"`
	Name                 string   `json:"name,omitempty"`
	DisplayName          string   `json:"displayName,omitempty"`
	Description          string   `json:"description,omitempty"`
	InfoUrls             []string `json:"infoUrls,omitempty"`
	Keywords             []string `json:"keywords,omitempty"`
	Occurrence           *Occurrence `json:"occurrence,omitempty"`
	BaseUri              string   `json:"baseUri,omitempty"`
	OffsetUri            string   `json:"offsetUri,omitempty"`
	RootDefinitionId     string   `json:"rootDefinitionId,omitempty"`
	CategoryId           string   `json:"categoryId,omitempty"`
	SettingUsage         string   `json:"settingUsage,omitempty"`
	UxBehavior           string   `json:"uxBehavior,omitempty"`
	Visibility           string   `json:"visibility,omitempty"`
	ReferredSettingInformationList []ReferredSettingInformation `json:"referredSettingInformationList,omitempty"`
	AccessTypes          string   `json:"accessTypes,omitempty"`
	Applicability        *Applicability `json:"applicability,omitempty"`
}

// Occurrence represents occurrence constraints for a setting
type Occurrence struct {
	MinDeviceOccurrence int `json:"minDeviceOccurrence,omitempty"`
	MaxDeviceOccurrence int `json:"maxDeviceOccurrence,omitempty"`
}

// ReferredSettingInformation represents referred setting information
type ReferredSettingInformation struct {
	SettingDefinitionId string `json:"settingDefinitionId,omitempty"`
}

// Applicability represents applicability information
type Applicability struct {
	Description  string   `json:"description,omitempty"`
	Platform     string   `json:"platform,omitempty"`
	DeviceMode   string   `json:"deviceMode,omitempty"`
	Technologies string   `json:"technologies,omitempty"`
}

// ScopeTag represents an Intune role scope tag
type ScopeTag struct {
	ODataType   string `json:"@odata.type,omitempty"`
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	IsBuiltIn   bool   `json:"isBuiltIn,omitempty"`
}

// AssignmentFilter represents an Intune assignment filter
type AssignmentFilter struct {
	ODataType                string   `json:"@odata.type,omitempty"`
	ID                       string   `json:"id,omitempty"`
	DisplayName              string   `json:"displayName"`
	Description              string   `json:"description,omitempty"`
	Platform                 string   `json:"platform"`
	Rule                     string   `json:"rule"`
	RoleScopeTags            []string `json:"roleScopeTags,omitempty"`
	CreatedDateTime          string   `json:"createdDateTime,omitempty"`
	LastModifiedDateTime     string   `json:"lastModifiedDateTime,omitempty"`
	AssignmentFilterManagementType string `json:"assignmentFilterManagementType,omitempty"`
}

// Intune API paths
const (
	// Settings Catalog
	PathSettingsCatalogPolicies     = "/deviceManagement/configurationPolicies"
	PathSettingsCatalogDefinitions  = "/deviceManagement/configurationPolicyTemplates"
	PathSettingsDefinitions         = "/deviceManagement/reusableSettings"

	// Compliance Policies
	PathCompliancePolicies          = "/deviceManagement/deviceCompliancePolicies"

	// Endpoint Security
	PathEndpointSecurityPolicies    = "/deviceManagement/intents"
	PathEndpointSecurityTemplates   = "/deviceManagement/templates"

	// Device Configuration
	PathDeviceConfigurations        = "/deviceManagement/deviceConfigurations"

	// Assignments
	PathAssignments                 = "/assignments"

	// Scope Tags
	PathScopeTags                   = "/deviceManagement/roleScopeTags"

	// Assignment Filters
	PathAssignmentFilters           = "/deviceManagement/assignmentFilters"
)

// CreateSettingsCatalogPolicy creates a new Settings Catalog policy
func (c *GraphClient) CreateSettingsCatalogPolicy(ctx context.Context, policy *SettingsCatalogPolicy) (*SettingsCatalogPolicy, error) {
	resp, err := c.Post(ctx, PathSettingsCatalogPolicies, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to create settings catalog policy: %w", err)
	}

	// Parse the response into a policy
	var created SettingsCatalogPolicy
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created policy: %w", err)
	}

	// The ID is in the response
	if created.ID == "" {
		created.ID = resp.ID
	}

	return &created, nil
}

// GetSettingsCatalogPolicy retrieves a Settings Catalog policy by ID
func (c *GraphClient) GetSettingsCatalogPolicy(ctx context.Context, id string) (*SettingsCatalogPolicy, error) {
	path := fmt.Sprintf("%s('%s')?$expand=settings", PathSettingsCatalogPolicies, id)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings catalog policy: %w", err)
	}

	// The response is the policy itself, not in Value
	respBytes, _ := json.Marshal(resp)
	var policy SettingsCatalogPolicy
	if err := json.Unmarshal(respBytes, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	if policy.ID == "" {
		policy.ID = resp.ID
	}

	return &policy, nil
}

// UpdateSettingsCatalogPolicy updates a Settings Catalog policy
func (c *GraphClient) UpdateSettingsCatalogPolicy(ctx context.Context, id string, policy *SettingsCatalogPolicy) (*SettingsCatalogPolicy, error) {
	path := fmt.Sprintf("%s('%s')", PathSettingsCatalogPolicies, id)
	_, err := c.Patch(ctx, path, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to update settings catalog policy: %w", err)
	}

	// Get the updated policy
	return c.GetSettingsCatalogPolicy(ctx, id)
}

// DeleteSettingsCatalogPolicy deletes a Settings Catalog policy
func (c *GraphClient) DeleteSettingsCatalogPolicy(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s('%s')", PathSettingsCatalogPolicies, id)
	return c.Delete(ctx, path)
}

// UpdateSettingsCatalogPolicySettings updates the settings of a Settings Catalog policy
func (c *GraphClient) UpdateSettingsCatalogPolicySettings(ctx context.Context, policyId string, settings []SettingsCatalogPolicySetting) error {
	path := fmt.Sprintf("%s('%s')/settings", PathSettingsCatalogPolicies, policyId)

	body := map[string]interface{}{
		"settings": settings,
	}

	_, err := c.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update settings catalog policy settings: %w", err)
	}

	return nil
}

// CreateCompliancePolicy creates a new compliance policy
func (c *GraphClient) CreateCompliancePolicy(ctx context.Context, policy *CompliancePolicy) (*CompliancePolicy, error) {
	resp, err := c.Post(ctx, PathCompliancePolicies, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to create compliance policy: %w", err)
	}

	respBytes, _ := json.Marshal(resp)
	var created CompliancePolicy
	if err := json.Unmarshal(respBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created policy: %w", err)
	}

	if created.ID == "" {
		created.ID = resp.ID
	}

	return &created, nil
}

// GetCompliancePolicy retrieves a compliance policy by ID
func (c *GraphClient) GetCompliancePolicy(ctx context.Context, id string) (*CompliancePolicy, error) {
	path := fmt.Sprintf("%s/%s", PathCompliancePolicies, id)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance policy: %w", err)
	}

	respBytes, _ := json.Marshal(resp)
	var policy CompliancePolicy
	if err := json.Unmarshal(respBytes, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	if policy.ID == "" {
		policy.ID = resp.ID
	}

	return &policy, nil
}

// UpdateCompliancePolicy updates a compliance policy
func (c *GraphClient) UpdateCompliancePolicy(ctx context.Context, id string, policy *CompliancePolicy) (*CompliancePolicy, error) {
	path := fmt.Sprintf("%s/%s", PathCompliancePolicies, id)
	_, err := c.Patch(ctx, path, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to update compliance policy: %w", err)
	}

	return c.GetCompliancePolicy(ctx, id)
}

// DeleteCompliancePolicy deletes a compliance policy
func (c *GraphClient) DeleteCompliancePolicy(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", PathCompliancePolicies, id)
	return c.Delete(ctx, path)
}

// GetPolicyAssignments retrieves assignments for a policy
func (c *GraphClient) GetPolicyAssignments(ctx context.Context, policyPath string, policyId string) ([]PolicyAssignment, error) {
	path := fmt.Sprintf("%s('%s')%s", policyPath, policyId, PathAssignments)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy assignments: %w", err)
	}

	var assignments []PolicyAssignment
	if resp.Value != nil {
		if err := json.Unmarshal(resp.Value, &assignments); err != nil {
			return nil, fmt.Errorf("failed to parse assignments: %w", err)
		}
	}

	return assignments, nil
}

// AssignPolicy assigns a policy to groups
func (c *GraphClient) AssignPolicy(ctx context.Context, policyPath string, policyId string, assignments []PolicyAssignment) error {
	path := fmt.Sprintf("%s('%s')/assign", policyPath, policyId)

	body := map[string]interface{}{
		"assignments": assignments,
	}

	_, err := c.Post(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to assign policy: %w", err)
	}

	return nil
}

// ListSettingDefinitions lists setting definitions for the Settings Catalog
func (c *GraphClient) ListSettingDefinitions(ctx context.Context, filter string) ([]SettingDefinition, error) {
	path := "/deviceManagement/configurationSettings"
	if filter != "" {
		path = fmt.Sprintf("%s?$filter=%s", path, url.QueryEscape(filter))
	}

	items, err := c.ListAll(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list setting definitions: %w", err)
	}

	var definitions []SettingDefinition
	for _, item := range items {
		var def SettingDefinition
		if err := json.Unmarshal(item, &def); err != nil {
			continue
		}
		definitions = append(definitions, def)
	}

	return definitions, nil
}

// GetSettingDefinition retrieves a specific setting definition
func (c *GraphClient) GetSettingDefinition(ctx context.Context, id string) (*SettingDefinition, error) {
	path := fmt.Sprintf("/deviceManagement/configurationSettings('%s')", id)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get setting definition: %w", err)
	}

	respBytes, _ := json.Marshal(resp)
	var def SettingDefinition
	if err := json.Unmarshal(respBytes, &def); err != nil {
		return nil, fmt.Errorf("failed to parse setting definition: %w", err)
	}

	return &def, nil
}

// ============================================================================
// Scope Tag Methods
// ============================================================================

// CreateScopeTag creates a new role scope tag
func (c *GraphClient) CreateScopeTag(ctx context.Context, tag *ScopeTag) (*ScopeTag, error) {
	resp, err := c.Post(ctx, PathScopeTags, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to create scope tag: %w", err)
	}

	respBytes, _ := json.Marshal(resp)
	var created ScopeTag
	if err := json.Unmarshal(respBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created scope tag: %w", err)
	}

	if created.ID == "" {
		created.ID = resp.ID
	}

	return &created, nil
}

// GetScopeTag retrieves a scope tag by ID
func (c *GraphClient) GetScopeTag(ctx context.Context, id string) (*ScopeTag, error) {
	path := fmt.Sprintf("%s/%s", PathScopeTags, id)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get scope tag: %w", err)
	}

	respBytes, _ := json.Marshal(resp)
	var tag ScopeTag
	if err := json.Unmarshal(respBytes, &tag); err != nil {
		return nil, fmt.Errorf("failed to parse scope tag: %w", err)
	}

	if tag.ID == "" {
		tag.ID = resp.ID
	}

	return &tag, nil
}

// UpdateScopeTag updates a scope tag
func (c *GraphClient) UpdateScopeTag(ctx context.Context, id string, tag *ScopeTag) (*ScopeTag, error) {
	path := fmt.Sprintf("%s/%s", PathScopeTags, id)
	_, err := c.Patch(ctx, path, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to update scope tag: %w", err)
	}

	return c.GetScopeTag(ctx, id)
}

// DeleteScopeTag deletes a scope tag
func (c *GraphClient) DeleteScopeTag(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", PathScopeTags, id)
	return c.Delete(ctx, path)
}

// ListScopeTags lists all scope tags
func (c *GraphClient) ListScopeTags(ctx context.Context) ([]ScopeTag, error) {
	items, err := c.ListAll(ctx, PathScopeTags)
	if err != nil {
		return nil, fmt.Errorf("failed to list scope tags: %w", err)
	}

	var tags []ScopeTag
	for _, item := range items {
		var tag ScopeTag
		if err := json.Unmarshal(item, &tag); err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// ============================================================================
// Assignment Filter Methods
// ============================================================================

// CreateAssignmentFilter creates a new assignment filter
func (c *GraphClient) CreateAssignmentFilter(ctx context.Context, filter *AssignmentFilter) (*AssignmentFilter, error) {
	resp, err := c.Post(ctx, PathAssignmentFilters, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to create assignment filter: %w", err)
	}

	respBytes, _ := json.Marshal(resp)
	var created AssignmentFilter
	if err := json.Unmarshal(respBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created assignment filter: %w", err)
	}

	if created.ID == "" {
		created.ID = resp.ID
	}

	return &created, nil
}

// GetAssignmentFilter retrieves an assignment filter by ID
func (c *GraphClient) GetAssignmentFilter(ctx context.Context, id string) (*AssignmentFilter, error) {
	path := fmt.Sprintf("%s/%s", PathAssignmentFilters, id)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment filter: %w", err)
	}

	respBytes, _ := json.Marshal(resp)
	var filter AssignmentFilter
	if err := json.Unmarshal(respBytes, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse assignment filter: %w", err)
	}

	if filter.ID == "" {
		filter.ID = resp.ID
	}

	return &filter, nil
}

// UpdateAssignmentFilter updates an assignment filter
func (c *GraphClient) UpdateAssignmentFilter(ctx context.Context, id string, filter *AssignmentFilter) (*AssignmentFilter, error) {
	path := fmt.Sprintf("%s/%s", PathAssignmentFilters, id)
	_, err := c.Patch(ctx, path, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to update assignment filter: %w", err)
	}

	return c.GetAssignmentFilter(ctx, id)
}

// DeleteAssignmentFilter deletes an assignment filter
func (c *GraphClient) DeleteAssignmentFilter(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", PathAssignmentFilters, id)
	return c.Delete(ctx, path)
}

// ListAssignmentFilters lists all assignment filters
func (c *GraphClient) ListAssignmentFilters(ctx context.Context) ([]AssignmentFilter, error) {
	items, err := c.ListAll(ctx, PathAssignmentFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to list assignment filters: %w", err)
	}

	var filters []AssignmentFilter
	for _, item := range items {
		var filter AssignmentFilter
		if err := json.Unmarshal(item, &filter); err != nil {
			continue
		}
		filters = append(filters, filter)
	}

	return filters, nil
}
