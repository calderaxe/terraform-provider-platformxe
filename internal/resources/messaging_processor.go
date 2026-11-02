// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &MessagingProcessorResource{}
var _ resource.ResourceWithConfigure = &MessagingProcessorResource{}

type MessagingProcessorResource struct {
	client *platformxe.Client
}

type MessagingProcessorResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	EmailEnabled            types.Bool   `tfsdk:"email_enabled"`
	EmailPreferredProviders types.List   `tfsdk:"email_preferred_providers"`
	EmailFromName           types.String `tfsdk:"email_from_name"`
	EmailReplyTo            types.String `tfsdk:"email_reply_to"`
	SmsEnabled              types.Bool   `tfsdk:"sms_enabled"`
	SmsPreferredProviders   types.List   `tfsdk:"sms_preferred_providers"`
	SmsDefaultRegion        types.String `tfsdk:"sms_default_region"`
	WhatsappEnabled         types.Bool   `tfsdk:"whatsapp_enabled"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
}

func NewMessagingProcessorResource() resource.Resource {
	return &MessagingProcessorResource{}
}

func (r *MessagingProcessorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_messaging_processor"
}

func (r *MessagingProcessorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the PlatformXe messaging processor configuration for email, SMS, and WhatsApp dispatch.",
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
				Description: "Whether the messaging processor is enabled. Defaults to true.",
			},
			"email_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether email dispatch is enabled. Defaults to true.",
			},
			"email_preferred_providers": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Preferred email provider fallback order.",
			},
			"email_from_name": schema.StringAttribute{
				Optional:    true,
				Description: "Default sender display name for emails.",
			},
			"email_reply_to": schema.StringAttribute{
				Optional:    true,
				Description: "Default reply-to address for emails.",
			},
			"sms_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether SMS dispatch is enabled. Defaults to true.",
			},
			"sms_preferred_providers": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Preferred SMS provider fallback order.",
			},
			"sms_default_region": schema.StringAttribute{
				Optional:    true,
				Description: "Default region code for SMS delivery.",
			},
			"whatsapp_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether WhatsApp dispatch is enabled. Defaults to false.",
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

func (r *MessagingProcessorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func messagingProcessorInput(ctx context.Context, plan *MessagingProcessorResourceModel, diags *diag.Diagnostics) map[string]interface{} {
	input := map[string]interface{}{
		"enabled":         plan.Enabled.ValueBool(),
		"emailEnabled":    plan.EmailEnabled.ValueBool(),
		"smsEnabled":      plan.SmsEnabled.ValueBool(),
		"whatsappEnabled": plan.WhatsappEnabled.ValueBool(),
	}

	if !plan.EmailPreferredProviders.IsNull() && !plan.EmailPreferredProviders.IsUnknown() {
		var providers []string
		diags.Append(plan.EmailPreferredProviders.ElementsAs(ctx, &providers, false)...)
		input["emailPreferredProviders"] = providers
	}

	if !plan.EmailFromName.IsNull() && !plan.EmailFromName.IsUnknown() {
		input["emailFromName"] = plan.EmailFromName.ValueString()
	}

	if !plan.EmailReplyTo.IsNull() && !plan.EmailReplyTo.IsUnknown() {
		input["emailReplyTo"] = plan.EmailReplyTo.ValueString()
	}

	if !plan.SmsPreferredProviders.IsNull() && !plan.SmsPreferredProviders.IsUnknown() {
		var providers []string
		diags.Append(plan.SmsPreferredProviders.ElementsAs(ctx, &providers, false)...)
		input["smsPreferredProviders"] = providers
	}

	if !plan.SmsDefaultRegion.IsNull() && !plan.SmsDefaultRegion.IsUnknown() {
		input["smsDefaultRegion"] = plan.SmsDefaultRegion.ValueString()
	}

	return input
}

func (r *MessagingProcessorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MessagingProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := messagingProcessorInput(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Messaging.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Messaging Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *MessagingProcessorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MessagingProcessorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Messaging.GetProcessor()
	if err != nil {
		resp.Diagnostics.AddError("Read Messaging Processor Failed", err.Error())
		return
	}

	state.Enabled = types.BoolValue(result.Enabled)
	mapProcessorConfigToComputedState(result, &state.ID, &state.CreatedAt, &state.UpdatedAt)

	cfg := result.Config
	if v, ok := cfg["emailEnabled"].(bool); ok {
		state.EmailEnabled = types.BoolValue(v)
	}
	if v, ok := cfg["emailPreferredProviders"].([]interface{}); ok {
		list, diags := types.ListValueFrom(ctx, types.StringType, interfaceSliceToStringSlice(v))
		resp.Diagnostics.Append(diags...)
		state.EmailPreferredProviders = list
	}
	if v, ok := cfg["emailFromName"].(string); ok {
		state.EmailFromName = types.StringValue(v)
	}
	if v, ok := cfg["emailReplyTo"].(string); ok {
		state.EmailReplyTo = types.StringValue(v)
	}
	if v, ok := cfg["smsEnabled"].(bool); ok {
		state.SmsEnabled = types.BoolValue(v)
	}
	if v, ok := cfg["smsPreferredProviders"].([]interface{}); ok {
		list, diags := types.ListValueFrom(ctx, types.StringType, interfaceSliceToStringSlice(v))
		resp.Diagnostics.Append(diags...)
		state.SmsPreferredProviders = list
	}
	if v, ok := cfg["smsDefaultRegion"].(string); ok {
		state.SmsDefaultRegion = types.StringValue(v)
	}
	if v, ok := cfg["whatsappEnabled"].(bool); ok {
		state.WhatsappEnabled = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *MessagingProcessorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MessagingProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := messagingProcessorInput(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Messaging.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Update Messaging Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *MessagingProcessorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.Messaging.UpdateProcessor(map[string]interface{}{"enabled": false})
	if err != nil {
		resp.Diagnostics.AddError("Disable Messaging Processor Failed", err.Error())
		return
	}
}
