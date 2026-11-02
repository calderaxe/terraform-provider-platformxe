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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &OcrProcessorResource{}
var _ resource.ResourceWithConfigure = &OcrProcessorResource{}

type OcrProcessorResource struct {
	client *platformxe.Client
}

type OcrProcessorResourceModel struct {
	ID                     types.String  `tfsdk:"id"`
	Enabled                types.Bool    `tfsdk:"enabled"`
	Provider               types.String  `tfsdk:"provider"`
	ConfidenceThreshold    types.Float64 `tfsdk:"confidence_threshold"`
	SupportedDocumentTypes types.List    `tfsdk:"supported_document_types"`
	Languages              types.List    `tfsdk:"languages"`
	CreatedAt              types.String  `tfsdk:"created_at"`
	UpdatedAt              types.String  `tfsdk:"updated_at"`
}

func NewOcrProcessorResource() resource.Resource {
	return &OcrProcessorResource{}
}

func (r *OcrProcessorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ocr_processor"
}

func (r *OcrProcessorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the PlatformXe OCR processor configuration for identity document verification.",
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
				Description: "Whether the OCR processor is enabled. Defaults to true.",
			},
			"provider": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("azure"),
				Description: "OCR provider. Defaults to \"azure\".",
			},
			"confidence_threshold": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(0.85),
				Description: "Minimum confidence score to accept OCR results. Defaults to 0.85.",
			},
			"supported_document_types": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of supported document types (e.g., \"passport\", \"drivers_license\", \"national_id\").",
			},
			"languages": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of supported languages for OCR processing.",
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

func (r *OcrProcessorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OcrProcessorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OcrProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":             plan.Enabled.ValueBool(),
		"provider":            plan.Provider.ValueString(),
		"confidenceThreshold": plan.ConfidenceThreshold.ValueFloat64(),
	}

	if !plan.SupportedDocumentTypes.IsNull() && !plan.SupportedDocumentTypes.IsUnknown() {
		var docTypes []string
		resp.Diagnostics.Append(plan.SupportedDocumentTypes.ElementsAs(ctx, &docTypes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["supportedDocumentTypes"] = docTypes
	}

	if !plan.Languages.IsNull() && !plan.Languages.IsUnknown() {
		var languages []string
		resp.Diagnostics.Append(plan.Languages.ElementsAs(ctx, &languages, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["languages"] = languages
	}

	result, err := r.client.Ocr.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Create OCR Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OcrProcessorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OcrProcessorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Ocr.GetProcessor()
	if err != nil {
		resp.Diagnostics.AddError("Read OCR Processor Failed", err.Error())
		return
	}

	state.Enabled = types.BoolValue(result.Enabled)
	mapProcessorConfigToComputedState(result, &state.ID, &state.CreatedAt, &state.UpdatedAt)

	cfg := result.Config
	if v, ok := cfg["provider"].(string); ok {
		state.Provider = types.StringValue(v)
	}
	if v, ok := cfg["confidenceThreshold"].(float64); ok {
		state.ConfidenceThreshold = types.Float64Value(v)
	}
	if v, ok := cfg["supportedDocumentTypes"].([]interface{}); ok {
		list, diags := types.ListValueFrom(ctx, types.StringType, interfaceSliceToStringSlice(v))
		resp.Diagnostics.Append(diags...)
		state.SupportedDocumentTypes = list
	}
	if v, ok := cfg["languages"].([]interface{}); ok {
		list, diags := types.ListValueFrom(ctx, types.StringType, interfaceSliceToStringSlice(v))
		resp.Diagnostics.Append(diags...)
		state.Languages = list
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *OcrProcessorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OcrProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":             plan.Enabled.ValueBool(),
		"provider":            plan.Provider.ValueString(),
		"confidenceThreshold": plan.ConfidenceThreshold.ValueFloat64(),
	}

	if !plan.SupportedDocumentTypes.IsNull() && !plan.SupportedDocumentTypes.IsUnknown() {
		var docTypes []string
		resp.Diagnostics.Append(plan.SupportedDocumentTypes.ElementsAs(ctx, &docTypes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["supportedDocumentTypes"] = docTypes
	}

	if !plan.Languages.IsNull() && !plan.Languages.IsUnknown() {
		var languages []string
		resp.Diagnostics.Append(plan.Languages.ElementsAs(ctx, &languages, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["languages"] = languages
	}

	result, err := r.client.Ocr.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Update OCR Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OcrProcessorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.Ocr.UpdateProcessor(map[string]interface{}{"enabled": false})
	if err != nil {
		resp.Diagnostics.AddError("Disable OCR Processor Failed", err.Error())
		return
	}
}

// -- Shared processor helpers -------------------------------------------------

func mapProcessorConfigToComputedState(result *platformxe.ProcessorConfig, id, createdAt, updatedAt *types.String) {
	if result.ID != "" {
		*id = types.StringValue(result.ID)
	}
	if result.CreatedAt != "" {
		*createdAt = types.StringValue(result.CreatedAt)
	}
	if result.UpdatedAt != "" {
		*updatedAt = types.StringValue(result.UpdatedAt)
	}
}

func interfaceSliceToStringSlice(in []interface{}) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
