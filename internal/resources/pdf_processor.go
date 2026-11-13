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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &PdfProcessorResource{}
var _ resource.ResourceWithConfigure = &PdfProcessorResource{}

type PdfProcessorResource struct {
	client *platformxe.Client
}

type PdfProcessorResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	DefaultTemplate  types.String `tfsdk:"default_template"`
	EnabledTemplates types.List   `tfsdk:"enabled_templates"`
	BrandingSource   types.String `tfsdk:"branding_source"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func NewPdfProcessorResource() resource.Resource {
	return &PdfProcessorResource{}
}

func (r *PdfProcessorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pdf_processor"
}

func (r *PdfProcessorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the PlatformXe PDF processor configuration for document generation.",
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
				Description: "Whether the PDF processor is enabled. Defaults to true.",
			},
			"default_template": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("standard"),
				Description: "Default PDF template. Defaults to \"standard\".",
			},
			"enabled_templates": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of enabled PDF template identifiers.",
			},
			"branding_source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("organization"),
				Description: "Source for branding assets. Defaults to \"organization\".",
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

func (r *PdfProcessorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PdfProcessorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PdfProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":         plan.Enabled.ValueBool(),
		"defaultTemplate": plan.DefaultTemplate.ValueString(),
		"brandingSource":  plan.BrandingSource.ValueString(),
	}

	if !plan.EnabledTemplates.IsNull() && !plan.EnabledTemplates.IsUnknown() {
		var templates []string
		resp.Diagnostics.Append(plan.EnabledTemplates.ElementsAs(ctx, &templates, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["enabledTemplates"] = templates
	}

	result, err := r.client.Pdf.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Create PDF Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PdfProcessorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PdfProcessorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Pdf.GetProcessor()
	if err != nil {
		resp.Diagnostics.AddError("Read PDF Processor Failed", err.Error())
		return
	}

	state.Enabled = types.BoolValue(result.Enabled)
	mapProcessorConfigToComputedState(result, &state.ID, &state.CreatedAt, &state.UpdatedAt)

	cfg := result.Config
	if v, ok := cfg["defaultTemplate"].(string); ok {
		state.DefaultTemplate = types.StringValue(v)
	}
	if v, ok := cfg["enabledTemplates"].([]interface{}); ok {
		list, diags := types.ListValueFrom(ctx, types.StringType, interfaceSliceToStringSlice(v))
		resp.Diagnostics.Append(diags...)
		state.EnabledTemplates = list
	}
	if v, ok := cfg["brandingSource"].(string); ok {
		state.BrandingSource = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PdfProcessorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PdfProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":         plan.Enabled.ValueBool(),
		"defaultTemplate": plan.DefaultTemplate.ValueString(),
		"brandingSource":  plan.BrandingSource.ValueString(),
	}

	if !plan.EnabledTemplates.IsNull() && !plan.EnabledTemplates.IsUnknown() {
		var templates []string
		resp.Diagnostics.Append(plan.EnabledTemplates.ElementsAs(ctx, &templates, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["enabledTemplates"] = templates
	}

	result, err := r.client.Pdf.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Update PDF Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PdfProcessorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.Pdf.UpdateProcessor(map[string]interface{}{"enabled": false})
	if err != nil {
		resp.Diagnostics.AddError("Disable PDF Processor Failed", err.Error())
		return
	}
}
