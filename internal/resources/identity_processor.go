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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &IdentityProcessorResource{}
var _ resource.ResourceWithConfigure = &IdentityProcessorResource{}

type IdentityProcessorResource struct {
	client *platformxe.Client
}

type IdentityProcessorResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	VerificationMethods types.List   `tfsdk:"verification_methods"`
	ConsentRequired     types.Bool   `tfsdk:"consent_required"`
	NdprCompliance      types.Bool   `tfsdk:"ndpr_compliance"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func NewIdentityProcessorResource() resource.Resource {
	return &IdentityProcessorResource{}
}

func (r *IdentityProcessorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_processor"
}

func (r *IdentityProcessorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the PlatformXe identity processor configuration for identity resolution and verification.",
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
				Description: "Whether the identity processor is enabled. Defaults to true.",
			},
			"verification_methods": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Allowed identity verification methods (e.g., \"bvn\", \"nin\", \"passport\", \"drivers_license\").",
			},
			"consent_required": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether explicit consent is required before identity lookups. Defaults to true.",
			},
			"ndpr_compliance": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether NDPR (Nigeria Data Protection Regulation) compliance checks are enforced. Defaults to true.",
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

func (r *IdentityProcessorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IdentityProcessorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IdentityProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":         plan.Enabled.ValueBool(),
		"consentRequired": plan.ConsentRequired.ValueBool(),
		"ndprCompliance":  plan.NdprCompliance.ValueBool(),
	}

	if !plan.VerificationMethods.IsNull() && !plan.VerificationMethods.IsUnknown() {
		var methods []string
		resp.Diagnostics.Append(plan.VerificationMethods.ElementsAs(ctx, &methods, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["verificationMethods"] = methods
	}

	result, err := r.client.Identity.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Identity Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *IdentityProcessorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IdentityProcessorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Identity.GetProcessor()
	if err != nil {
		resp.Diagnostics.AddError("Read Identity Processor Failed", err.Error())
		return
	}

	state.Enabled = types.BoolValue(result.Enabled)
	mapProcessorConfigToComputedState(result, &state.ID, &state.CreatedAt, &state.UpdatedAt)

	cfg := result.Config
	if v, ok := cfg["verificationMethods"].([]interface{}); ok {
		list, diags := types.ListValueFrom(ctx, types.StringType, interfaceSliceToStringSlice(v))
		resp.Diagnostics.Append(diags...)
		state.VerificationMethods = list
	}
	if v, ok := cfg["consentRequired"].(bool); ok {
		state.ConsentRequired = types.BoolValue(v)
	}
	if v, ok := cfg["ndprCompliance"].(bool); ok {
		state.NdprCompliance = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *IdentityProcessorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IdentityProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":         plan.Enabled.ValueBool(),
		"consentRequired": plan.ConsentRequired.ValueBool(),
		"ndprCompliance":  plan.NdprCompliance.ValueBool(),
	}

	if !plan.VerificationMethods.IsNull() && !plan.VerificationMethods.IsUnknown() {
		var methods []string
		resp.Diagnostics.Append(plan.VerificationMethods.ElementsAs(ctx, &methods, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["verificationMethods"] = methods
	}

	result, err := r.client.Identity.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Update Identity Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *IdentityProcessorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.Identity.UpdateProcessor(map[string]interface{}{"enabled": false})
	if err != nil {
		resp.Diagnostics.AddError("Disable Identity Processor Failed", err.Error())
		return
	}
}
