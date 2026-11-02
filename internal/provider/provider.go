// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
	"github.com/calderaxe/terraform-provider-platformxe/internal/datasources"
	"github.com/calderaxe/terraform-provider-platformxe/internal/resources"
)

var _ provider.Provider = &PlatformXeProvider{}

type PlatformXeProvider struct{}

type PlatformXeProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New() provider.Provider {
	return &PlatformXeProvider{}
}

func (p *PlatformXeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "platformxe"
}

func (p *PlatformXeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for PlatformXe — manage roles, policies, webhooks, workflows, and more as infrastructure-as-code.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "PlatformXe API key. Can also be set via PLATFORMXE_API_KEY env var.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "PlatformXe API base URL. Defaults to https://platformxe.com.",
				Optional:    true,
			},
		},
	}
}

func (p *PlatformXeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config PlatformXeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("PLATFORMXE_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	baseURL := "https://platformxe.com"
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError("Missing API Key", "api_key must be set in provider config or PLATFORMXE_API_KEY env var")
		return
	}

	client := platformxe.NewClient(platformxe.ClientConfig{
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Timeout:  30,
		Retries:  2,
		FailOpen: false,
	})

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PlatformXeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewPermissionsRoleResource,
		resources.NewPermissionsPolicyResource,
		resources.NewPermissionsOverrideResource,
		resources.NewPermissionsModuleResource,
		resources.NewPermissionsFederationGroupResource,
		resources.NewPermissionsFederationMemberResource,
		resources.NewWebhooksEndpointResource,
		resources.NewTemplatesTemplateResource,
		resources.NewWorkflowsTriggerResource,
		resources.NewDomainsSendingResource,
		resources.NewEventsSubscriptionResource,
		resources.NewThreadsChannelResource,
		resources.NewOcrProcessorResource,
		resources.NewPdfProcessorResource,
		resources.NewQrProcessorResource,
		resources.NewMessagingProcessorResource,
		resources.NewStorageProcessorResource,
		resources.NewExportsProcessorResource,
		resources.NewIdentityProcessorResource,
		// v1.1.0
		resources.NewFraudRuleResource,
		resources.NewFraudScreeningListResource,
		// v1.2.0 — Phase 9A tenant custom events
		resources.NewCustomEventResource,
		// v1.3.0 — Phase 9C tenant custom events marketplace
		resources.NewMarketplaceListingResource,
		// v1.4.0 — Phase 9D tenant custom event federation push (ENTERPRISE)
		resources.NewEventFederationGroupResource,
		resources.NewEventFederationPushResource,
		// v1.5.0 — Pattern 3 external_webhook peers (ENTERPRISE)
		resources.NewEventFederationExternalPeerResource,
	}
}

func (p *PlatformXeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewTenantDataSource,
		datasources.NewPermissionsModulesDataSource,
		datasources.NewIdentityProvidersDataSource,
		datasources.NewFraudTermsDataSource,
		datasources.NewThreadsEscalationConfigDataSource,
	}
}
