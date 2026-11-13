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

var _ datasource.DataSource = &ModulesDataSource{}
var _ datasource.DataSourceWithConfigure = &ModulesDataSource{}

// ModulesDataSource reads all registered permission modules.
// Useful for referencing available modules when defining roles and capabilities.
type ModulesDataSource struct {
	client *platformxe.Client
}

type ModuleItem struct {
	ID          types.String `tfsdk:"id"`
	App         types.String `tfsdk:"app"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type ModulesDataSourceModel struct {
	Modules []ModuleItem `tfsdk:"modules"`
}

func NewPermissionsModulesDataSource() datasource.DataSource {
	return &ModulesDataSource{}
}

func (d *ModulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_modules"
}

func (d *ModulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all registered permission modules across the PlatformXe ecosystem. Use this to discover available modules when configuring roles and capabilities.",
		Attributes: map[string]schema.Attribute{
			"modules": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of registered permission modules.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Module ID (e.g., LT:BOOKINGS).",
						},
						"app": schema.StringAttribute{
							Computed:    true,
							Description: "Source app that registered this module.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Module display name.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Module description.",
						},
					},
				},
			},
		},
	}
}

func (d *ModulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ModulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Permissions.ListModules()
	if err != nil {
		resp.Diagnostics.AddError("Read Modules Failed", err.Error())
		return
	}

	var state ModulesDataSourceModel

	// GET /api/v1/permissions/modules returns PermissionModuleSummary
	// (key, name, description, category, actions): `id` is the module key and
	// `app` is its category — the app that registered it.
	state.Modules = make([]ModuleItem, 0, len(result.Modules))
	for _, m := range result.Modules {
		state.Modules = append(state.Modules, ModuleItem{
			ID:          types.StringValue(m.Key),
			App:         types.StringValue(m.Category),
			Name:        types.StringValue(m.Name),
			Description: types.StringValue(m.Description),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
