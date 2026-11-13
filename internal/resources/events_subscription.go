// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &SubscriptionResource{}
var _ resource.ResourceWithConfigure = &SubscriptionResource{}
var _ resource.ResourceWithImportState = &SubscriptionResource{}

// ImportState imports a subscription by its ID. The signing secret is only
// returned at creation, so an imported subscription has an empty secret.
func (r *SubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type SubscriptionResource struct {
	client *platformxe.Client
}

type SubscriptionResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	EventTypes types.List   `tfsdk:"event_types"`
	WebhookURL types.String `tfsdk:"webhook_url"`
	IsActive   types.Bool   `tfsdk:"is_active"`
	Secret     types.String `tfsdk:"secret"`
}

// subscriptionStatus maps is_active onto the API's status. FAILING is set by
// PlatformXe after delivery failures; sending ACTIVE resets it.
func subscriptionStatus(isActive bool) string {
	if isActive {
		return "ACTIVE"
	}
	return "PAUSED"
}

func NewEventsSubscriptionResource() resource.Resource {
	return &SubscriptionResource{}
}

func (r *SubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_events_subscription"
}

func (r *SubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe event subscription.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Subscription identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Subscription name (at least 3 characters).",
			},
			"event_types": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of event types to subscribe to.",
			},
			"webhook_url": schema.StringAttribute{
				Required:    true,
				Description: "URL to receive event notifications.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the subscription is active (status ACTIVE) or paused (PAUSED). Defaults to true.",
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Signing secret for the delivered events. Returned once, at creation — store state encrypted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var eventTypes []string
	resp.Diagnostics.Append(plan.EventTypes.ElementsAs(ctx, &eventTypes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"eventTypes": eventTypes,
		"webhookUrl": plan.WebhookURL.ValueString(),
	}

	result, err := r.client.Events.CreateSubscription(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Subscription Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.Secret = types.StringValue(result.Secret)

	// A new subscription is always ACTIVE; pausing it is a second call.
	if !plan.IsActive.ValueBool() {
		if _, err := r.client.Events.UpdateSubscription(result.ID, map[string]interface{}{"status": "PAUSED"}); err != nil {
			// Keep the created subscription in state so it is not orphaned;
			// the next apply retries the pause.
			plan.IsActive = types.BoolValue(true)
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			resp.Diagnostics.AddError("Pause Subscription Failed", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Events.GetSubscription(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Subscription Failed", err.Error())
		return
	}

	state.Name = types.StringValue(result.Name)
	state.WebhookURL = types.StringValue(result.WebhookURL)
	// FAILING is still a live subscription; only PAUSED is inactive.
	state.IsActive = types.BoolValue(result.Status != "PAUSED")

	listVal, diags := types.ListValueFrom(ctx, types.StringType, result.EventTypes)
	resp.Diagnostics.Append(diags...)
	state.EventTypes = listVal

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var eventTypes []string
	resp.Diagnostics.Append(plan.EventTypes.ElementsAs(ctx, &eventTypes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"eventTypes": eventTypes,
		"webhookUrl": plan.WebhookURL.ValueString(),
		"status":     subscriptionStatus(plan.IsActive.ValueBool()),
	}

	_, err := r.client.Events.UpdateSubscription(plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Update Subscription Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Events.DeleteSubscription(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Subscription Failed", err.Error())
		return
	}
}
