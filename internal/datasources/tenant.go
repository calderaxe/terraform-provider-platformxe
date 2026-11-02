// =============================================================================
// © 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ datasource.DataSource = &TenantDataSource{}
var _ datasource.DataSourceWithConfigure = &TenantDataSource{}

// TenantDataSource reads your PlatformXe tenant details.
// The tenant is identified by the API key used in the provider configuration — no ID required.
type TenantDataSource struct {
	client *platformxe.Client
}

type TenantDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	Plan         types.String `tfsdk:"plan"`
	BillingEmail types.String `tfsdk:"billing_email"`
	Region       types.String `tfsdk:"region"`
	IsActive     types.Bool   `tfsdk:"is_active"`
}

func NewTenantDataSource() datasource.DataSource {
	return &TenantDataSource{}
}

func (d *TenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *TenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads your PlatformXe tenant details. The tenant is identified by the API key used in the provider configuration — no ID required.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant organization ID. The API key on the provider block identifies your tenant.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant display name.",
			},
			"slug": schema.StringAttribute{
				Computed:    true,
				Description: "URL-safe slug.",
			},
			"plan": schema.StringAttribute{
				Computed:    true,
				Description: "Subscription plan: FREE, BASIC, PRO, ENTERPRISE.",
			},
			"billing_email": schema.StringAttribute{
				Computed:    true,
				Description: "Billing contact email.",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "Region code (e.g., NG).",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the tenant is active.",
			},
		},
	}
}

func (d *TenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*platformxe.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected *platformxe.Client")
		return
	}
	d.client = client
}

func (d *TenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config TenantDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The /api/v1/tenant/me endpoint resolves the tenant from the API key context
	result, err := d.client.Get("/api/v1/tenant/me", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Tenant Failed", err.Error())
		return
	}

	if id, ok := result["id"].(string); ok {
		config.ID = types.StringValue(id)
	}
	if name, ok := result["name"].(string); ok {
		config.Name = types.StringValue(name)
	}
	if slug, ok := result["slug"].(string); ok {
		config.Slug = types.StringValue(slug)
	}
	if plan, ok := result["plan"].(string); ok {
		config.Plan = types.StringValue(plan)
	}
	if email, ok := result["billingEmail"].(string); ok {
		config.BillingEmail = types.StringValue(email)
	}
	if region, ok := result["region"].(string); ok {
		config.Region = types.StringValue(region)
	}
	if active, ok := result["isActive"].(bool); ok {
		config.IsActive = types.BoolValue(active)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
