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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

// platformxe_event_federation_external_peer — declarative state for a
// Pattern 3 external (non-tenant) peer on a Custom Event Federation group
// (added in PlatformXe 1.5.0). The owner adds an arbitrary HTTPS endpoint
// as a member of the group; relays are signed with a one-time HMAC secret
// surfaced at create time.
//
// IMPORTANT: the `secret` attribute is only populated on Create. The API
// does not echo it back on Read — once stored to state, it is immutable
// for the lifetime of the resource. ALL state files containing this
// resource MUST be encrypted at rest. Rotate by destroying and re-creating
// the resource.
//
// Maps to:
//   POST   /api/v1/events/custom/federation/groups/[gid]/external-peers (create)
//   GET    /api/v1/events/custom/federation/groups/[gid]                (read; locate by id)
//   DELETE /api/v1/events/custom/federation/external-peers/[id]         (delete = archive)

var _ resource.Resource = &EventFederationExternalPeerResource{}
var _ resource.ResourceWithConfigure = &EventFederationExternalPeerResource{}

type EventFederationExternalPeerResource struct {
	client *platformxe.Client
}

type EventFederationExternalPeerResourceModel struct {
	ID         types.String `tfsdk:"id"`
	GroupID    types.String `tfsdk:"group_id"`
	Label      types.String `tfsdk:"label"`
	WebhookURL types.String `tfsdk:"webhook_url"`
	Headers    types.Map    `tfsdk:"headers"`

	// Computed (server-assigned)
	Secret              types.String `tfsdk:"secret"`
	Status              types.String `tfsdk:"status"`
	ExternalHeaderNames types.List   `tfsdk:"external_header_names"`
	InvitedBy           types.String `tfsdk:"invited_by"`
	InvitedAt           types.String `tfsdk:"invited_at"`
}

func NewEventFederationExternalPeerResource() resource.Resource {
	return &EventFederationExternalPeerResource{}
}

func (r *EventFederationExternalPeerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_federation_external_peer"
}

func (r *EventFederationExternalPeerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pattern 3 external_webhook peer on a Custom Event Federation group (ENTERPRISE only). The peer is an arbitrary HTTPS endpoint that receives signed event relays. The HMAC `secret` is shown ONCE at create time and stored in Terraform state — encrypt your state at rest.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned external peer id (cefm_…).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "Federation group id (cefg_…) the peer belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable label for the external peer (e.g. \"Booking.com\").",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"webhook_url": schema.StringAttribute{
				Required:    true,
				Description: "HTTPS endpoint that receives the signed relay POSTs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"headers": schema.MapAttribute{
				Optional:      true,
				Sensitive:     true,
				ElementType:   types.StringType,
				Description:   "Optional static headers replayed on every relay (e.g. an Authorization Bearer token). Values are stored encrypted server-side and treated as sensitive in state. Only the header NAMES are echoed back on read; values cannot be changed without replacing the resource.",
				PlanModifiers: []planmodifier.Map{},
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "One-time HMAC-SHA256 signing secret (whsec_… prefix). Surfaced ONLY at create time — encrypt your state at rest. Used by the receiving endpoint to verify the X-Platformxe-Signature header on inbound webhooks. Rotate by destroying and re-creating the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Lifecycle state — one of pending|accepted|paused|removed. External peers are auto-accepted on create.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_header_names": schema.ListAttribute{
				Computed:      true,
				ElementType:   types.StringType,
				Description:   "Header names (without values) that the platform replays on each relay. Mirrors the headers map keys — values are NOT echoed for security.",
				PlanModifiers: []planmodifier.List{},
			},
			"invited_by": schema.StringAttribute{
				Computed:    true,
				Description: "Caller service or user that added the external peer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"invited_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO-8601 timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *EventFederationExternalPeerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EventFederationExternalPeerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EventFederationExternalPeerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := platformxe.AddEventFederationExternalPeerInput{
		Label:      plan.Label.ValueString(),
		WebhookURL: plan.WebhookURL.ValueString(),
	}

	if !plan.Headers.IsNull() && !plan.Headers.IsUnknown() {
		headers := map[string]string{}
		resp.Diagnostics.Append(plan.Headers.ElementsAs(ctx, &headers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(headers) > 0 {
			input.Headers = headers
		}
	}

	result, err := r.client.Events.Custom.Federation.AddExternalPeer(plan.GroupID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Add Event Federation External Peer Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.Peer.ID)
	plan.Secret = types.StringValue(result.Secret)
	plan.Status = types.StringValue(string(result.Peer.Status))
	plan.InvitedBy = types.StringValue(result.Peer.InvitedBy)
	plan.InvitedAt = types.StringValue(result.Peer.InvitedAt)

	headerNames, diags := types.ListValueFrom(ctx, types.StringType, result.Peer.ExternalWebhookHeaderNames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ExternalHeaderNames = headerNames

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventFederationExternalPeerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EventFederationExternalPeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// External peers are listed via the group detail endpoint. Locate by
	// id; if missing or status='removed', drop state so Terraform plans a
	// re-create.
	detail, err := r.client.Events.Custom.Federation.GetGroup(state.GroupID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var found *platformxe.EventFederationMemberSummary
	for i := range detail.Members {
		if detail.Members[i].ID == state.ID.ValueString() {
			found = &detail.Members[i]
			break
		}
	}
	if found == nil || found.Status == platformxe.EventFederationMemberStatusRemoved {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Status = types.StringValue(string(found.Status))
	state.InvitedBy = types.StringValue(found.InvitedBy)
	state.InvitedAt = types.StringValue(found.InvitedAt)
	if found.ExternalWebhookLabel != nil {
		state.Label = types.StringValue(*found.ExternalWebhookLabel)
	}
	if found.ExternalWebhookURL != nil {
		state.WebhookURL = types.StringValue(*found.ExternalWebhookURL)
	}

	headerNames, diags := types.ListValueFrom(ctx, types.StringType, found.ExternalWebhookHeaderNames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ExternalHeaderNames = headerNames

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EventFederationExternalPeerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All shape-affecting attributes have RequiresReplace; nothing to do.
	var plan EventFederationExternalPeerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventFederationExternalPeerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EventFederationExternalPeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Events.Custom.Federation.RemoveExternalPeer(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Remove Event Federation External Peer Failed", err.Error())
		return
	}
}
