// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &ExportsProcessorResource{}
var _ resource.ResourceWithConfigure = &ExportsProcessorResource{}

type ExportsProcessorResource struct {
	client *platformxe.Client
}

type ExportsProcessorResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	AllowedFormats   types.List   `tfsdk:"allowed_formats"`
	MaxRowsPerExport types.Int64  `tfsdk:"max_rows_per_export"`
	RetentionDays    types.Int64  `tfsdk:"retention_days"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func NewExportsProcessorResource() resource.Resource {
	return &ExportsProcessorResource{}
}

func (r *ExportsProcessorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_exports_processor"
}

func (r *ExportsProcessorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the PlatformXe exports processor configuration for data export jobs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Processor identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the exports processor is enabled. Defaults to true.",
			},
			"allowed_formats": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Allowed export formats. Defaults to [\"csv\", \"json\"].",
			},
			"max_rows_per_export": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100000),
				Description: "Maximum number of rows per export job. Defaults to 100000.",
			},
			"retention_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "Number of days to retain export files. Defaults to 30.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the processor was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the processor was last updated.",
			},
		},
	}
}

func (r *ExportsProcessorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*platformxe.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *platformxe.Client")
		return
	}
	r.client = client
}

func (r *ExportsProcessorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ExportsProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":          plan.Enabled.ValueBool(),
		"maxRowsPerExport": plan.MaxRowsPerExport.ValueInt64(),
		"retentionDays":    plan.RetentionDays.ValueInt64(),
	}

	if !plan.AllowedFormats.IsNull() && !plan.AllowedFormats.IsUnknown() {
		var formats []string
		resp.Diagnostics.Append(plan.AllowedFormats.ElementsAs(ctx, &formats, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["allowedFormats"] = formats
	}

	result, err := r.client.Exports.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Exports Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ExportsProcessorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ExportsProcessorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Exports.GetProcessor()
	if err != nil {
		resp.Diagnostics.AddError("Read Exports Processor Failed", err.Error())
		return
	}

	state.Enabled = types.BoolValue(result.Enabled)
	mapProcessorConfigToComputedState(result, &state.ID, &state.CreatedAt, &state.UpdatedAt)

	cfg := result.Config
	if v, ok := cfg["allowedFormats"].([]interface{}); ok {
		list, diags := types.ListValueFrom(ctx, types.StringType, interfaceSliceToStringSlice(v))
		resp.Diagnostics.Append(diags...)
		state.AllowedFormats = list
	}
	if v, ok := cfg["maxRowsPerExport"].(float64); ok {
		state.MaxRowsPerExport = types.Int64Value(int64(v))
	}
	if v, ok := cfg["retentionDays"].(float64); ok {
		state.RetentionDays = types.Int64Value(int64(v))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ExportsProcessorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ExportsProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":          plan.Enabled.ValueBool(),
		"maxRowsPerExport": plan.MaxRowsPerExport.ValueInt64(),
		"retentionDays":    plan.RetentionDays.ValueInt64(),
	}

	if !plan.AllowedFormats.IsNull() && !plan.AllowedFormats.IsUnknown() {
		var formats []string
		resp.Diagnostics.Append(plan.AllowedFormats.ElementsAs(ctx, &formats, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["allowedFormats"] = formats
	}

	result, err := r.client.Exports.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Update Exports Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ExportsProcessorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.Exports.UpdateProcessor(map[string]interface{}{"enabled": false})
	if err != nil {
		resp.Diagnostics.AddError("Disable Exports Processor Failed", err.Error())
		return
	}
}
