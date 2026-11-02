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

// platformxe_event_federation_group — declarative state for a Custom
// Event Federation group (Phase 9D, ENTERPRISE only). Owners create
// groups; members are managed separately (out-of-band today;
// platformxe_event_federation_member is roadmapped for v2 once the
// dual-sided invite/accept lifecycle has a clean Terraform model).
//
// Naming: this is the *Custom Event* federation, distinct from the
// platformxe_permissions_federation_group resource which models the
// v1.x.x admin permissions cross-app federation.
//
// Maps to:
//   POST   /api/v1/events/custom/federation/groups        (create)
//   GET    /api/v1/events/custom/federation/groups/[id]   (read)
//   DELETE /api/v1/events/custom/federation/groups/[id]   (delete = archive)

var _ resource.Resource = &EventFederationGroupResource{}
var _ resource.ResourceWithConfigure = &EventFederationGroupResource{}

type EventFederationGroupResource struct {
	client *platformxe.Client
}

type EventFederationGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`

	// Computed (server-assigned)
	OwnerOrganizationID types.String `tfsdk:"owner_organization_id"`
	CreatedBy           types.String `tfsdk:"created_by"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	ArchivedAt          types.String `tfsdk:"archived_at"`
}

func NewEventFederationGroupResource() resource.Resource {
	return &EventFederationGroupResource{}
}

func (r *EventFederationGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_federation_group"
}

func (r *EventFederationGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Custom Event Federation group (ENTERPRISE only). Distinct from platformxe_permissions_federation_group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned group id (cefg_…).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Owner-scope-unique label (3–80 characters).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Free-text description of the group's purpose.",
			},
			"owner_organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "The organization id that owns the group (= the API key's org).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "Caller service or user that created the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO-8601 timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO-8601 timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"archived_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO-8601 timestamp once the group is archived (null while active).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *EventFederationGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EventFederationGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EventFederationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := platformxe.CreateEventFederationGroupInput{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		v := plan.Description.ValueString()
		input.Description = &v
	}

	result, err := r.client.Events.Custom.Federation.CreateGroup(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Event Federation Group Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.OwnerOrganizationID = types.StringValue(result.OwnerOrganizationID)
	plan.CreatedBy = types.StringValue(result.CreatedBy)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)
	if result.ArchivedAt != nil {
		plan.ArchivedAt = types.StringValue(*result.ArchivedAt)
	} else {
		plan.ArchivedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventFederationGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EventFederationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := r.client.Events.Custom.Federation.GetGroup(state.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(detail.Name)
	if detail.Description != nil {
		state.Description = types.StringValue(*detail.Description)
	}
	state.OwnerOrganizationID = types.StringValue(detail.OwnerOrganizationID)
	state.CreatedBy = types.StringValue(detail.CreatedBy)
	state.CreatedAt = types.StringValue(detail.CreatedAt)
	state.UpdatedAt = types.StringValue(detail.UpdatedAt)
	if detail.ArchivedAt != nil {
		state.ArchivedAt = types.StringValue(*detail.ArchivedAt)
	} else {
		state.ArchivedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EventFederationGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// `name` has RequiresReplace; description/etc. are state-only no-ops
	// (the API doesn't yet expose a PATCH endpoint).
	var plan EventFederationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventFederationGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EventFederationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Events.Custom.Federation.ArchiveGroup(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Archive Event Federation Group Failed", err.Error())
		return
	}
}
