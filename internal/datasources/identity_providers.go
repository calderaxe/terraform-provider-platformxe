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

var _ datasource.DataSource = &IdentityProvidersDataSource{}
var _ datasource.DataSourceWithConfigure = &IdentityProvidersDataSource{}

// IdentityProvidersDataSource reads the health status of identity resolution providers.
// Useful for monitoring and conditional logic in Terraform configurations.
type IdentityProvidersDataSource struct {
	client *platformxe.Client
}

type ProviderItem struct {
	Name    types.String `tfsdk:"name"`
	Healthy types.Bool   `tfsdk:"healthy"`
}

type IdentityProvidersDataSourceModel struct {
	Providers []ProviderItem `tfsdk:"providers"`
}

func NewIdentityProvidersDataSource() datasource.DataSource {
	return &IdentityProvidersDataSource{}
}

func (d *IdentityProvidersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_providers"
}

func (d *IdentityProvidersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists identity resolution providers and their health status. Use this to check provider availability before configuring identity verification workflows.",
		Attributes: map[string]schema.Attribute{
			"providers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of identity providers with health status.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Provider name (e.g., smile-id, nimc, dojah).",
						},
						"healthy": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the provider is currently healthy and accepting requests.",
						},
					},
				},
			},
		},
	}
}

func (d *IdentityProvidersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IdentityProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Identity.Providers()
	if err != nil {
		resp.Diagnostics.AddError("Read Identity Providers Failed", err.Error())
		return
	}

	var state IdentityProvidersDataSourceModel

	state.Providers = make([]ProviderItem, 0, len(result.Providers))
	for _, p := range result.Providers {
		state.Providers = append(state.Providers, ProviderItem{
			Name:    types.StringValue(p.Name),
			Healthy: types.BoolValue(p.Healthy),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
