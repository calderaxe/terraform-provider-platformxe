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

// platformxe_event_federation_push — declarative state for a per-version
// event push declaration on a Custom Event Federation group (Phase 9D).
// Owners declare which of their registered events flow to a group's
// accepted members.
//
// Maps to:
//   POST   /api/v1/events/custom/federation/groups/[gid]/pushes  (create)
//   GET    /api/v1/events/custom/federation/groups/[gid]/pushes  (read; filter to id)
//   DELETE /api/v1/events/custom/federation/pushes/[id]          (delete = undeclare)

var _ resource.Resource = &EventFederationPushResource{}
var _ resource.ResourceWithConfigure = &EventFederationPushResource{}

type EventFederationPushResource struct {
	client *platformxe.Client
}

type EventFederationPushResourceModel struct {
	ID             types.String `tfsdk:"id"`
	GroupID        types.String `tfsdk:"group_id"`
	RegistrationID types.String `tfsdk:"registration_id"`

	// Computed (server-assigned, snapshot at push-declaration time)
	SourceOrganizationID types.String `tfsdk:"source_organization_id"`
	Namespace            types.String `tfsdk:"namespace"`
	Name                 types.String `tfsdk:"name"`
	Version              types.String `tfsdk:"version"`
	SourceCanonicalName  types.String `tfsdk:"source_canonical_name"`
	IsActive             types.Bool   `tfsdk:"is_active"`
	PushedBy             types.String `tfsdk:"pushed_by"`
	PushedAt             types.String `tfsdk:"pushed_at"`
}

func NewEventFederationPushResource() resource.Resource {
	return &EventFederationPushResource{}
}

func (r *EventFederationPushResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_federation_push"
}

func (r *EventFederationPushResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a per-version event push declaration on a Custom Event Federation group. ENTERPRISE only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned push id (cefp_…).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "Federation group id (cefg_…) the push belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registration_id": schema.StringAttribute{
				Required:    true,
				Description: "Source custom-event registration id (cer_…). Must be owned by the calling org and status='published'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "Owner organization id (cached at push time).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Computed:    true,
				Description: "Namespace cached at push time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Event name cached at push time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Semver cached at push time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_canonical_name": schema.StringAttribute{
				Computed:    true,
				Description: "TENANT_CUSTOM:<orgId>:<ns>.<name>@<version> at push time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_active": schema.BoolAttribute{
				Computed:      true,
				Description:   "True while the push is active. False after delete (audit history preserved).",
				PlanModifiers: []planmodifier.Bool{},
			},
			"pushed_by": schema.StringAttribute{
				Computed:    true,
				Description: "Caller service or user that declared the push.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pushed_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO-8601 timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *EventFederationPushResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EventFederationPushResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EventFederationPushResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := platformxe.DeclareEventFederationPushInput{
		RegistrationID: plan.RegistrationID.ValueString(),
	}

	result, err := r.client.Events.Custom.Federation.DeclarePush(plan.GroupID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Declare Event Federation Push Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.SourceOrganizationID = types.StringValue(result.SourceOrganizationID)
	plan.Namespace = types.StringValue(result.Namespace)
	plan.Name = types.StringValue(result.Name)
	plan.Version = types.StringValue(result.Version)
	plan.SourceCanonicalName = types.StringValue(result.SourceCanonicalName)
	plan.IsActive = types.BoolValue(result.IsActive)
	plan.PushedBy = types.StringValue(result.PushedBy)
	plan.PushedAt = types.StringValue(result.PushedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventFederationPushResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EventFederationPushResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The push surface only exposes a list endpoint scoped to a group.
	// Fetch all and locate by id; if missing, the push has been
	// undeclared (or the caller is no longer a member) — drop state.
	includeInactive := true
	pushes, err := r.client.Events.Custom.Federation.ListPushes(
		state.GroupID.ValueString(),
		&platformxe.ListEventFederationPushesParams{IncludeInactive: &includeInactive},
	)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	var found *platformxe.EventFederationPushSummary
	for i := range pushes.Pushes {
		if pushes.Pushes[i].ID == state.ID.ValueString() {
			found = &pushes.Pushes[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.SourceOrganizationID = types.StringValue(found.SourceOrganizationID)
	state.Namespace = types.StringValue(found.Namespace)
	state.Name = types.StringValue(found.Name)
	state.Version = types.StringValue(found.Version)
	state.SourceCanonicalName = types.StringValue(found.SourceCanonicalName)
	state.IsActive = types.BoolValue(found.IsActive)
	state.PushedBy = types.StringValue(found.PushedBy)
	state.PushedAt = types.StringValue(found.PushedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EventFederationPushResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All shape-affecting attributes have RequiresReplace.
	var plan EventFederationPushResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventFederationPushResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EventFederationPushResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Events.Custom.Federation.UndeclarePush(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Undeclare Event Federation Push Failed", err.Error())
		return
	}
}
