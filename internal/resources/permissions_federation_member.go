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

var _ resource.Resource = &FederationMemberResource{}
var _ resource.ResourceWithConfigure = &FederationMemberResource{}

type FederationMemberResource struct {
	client *platformxe.Client
}

type FederationMemberResourceModel struct {
	ID             types.String `tfsdk:"id"`
	GroupID        types.String `tfsdk:"group_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Prefix         types.String `tfsdk:"prefix"`
}

func NewPermissionsFederationMemberResource() resource.Resource {
	return &FederationMemberResource{}
}

func (r *FederationMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_federation_member"
}

func (r *FederationMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a member within a PlatformXe federation group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Federation member identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The federation group ID this member belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				Required:    true,
				Description: "The organization ID of the member.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.StringAttribute{
				Required:    true,
				Description: "Short prefix for the member within the federation (e.g. LT, CH).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *FederationMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FederationMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FederationMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"organizationId": plan.OrganizationID.ValueString(),
		"prefix":         plan.Prefix.ValueString(),
	}

	result, err := r.client.Permissions.AddFederationMember(plan.GroupID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Create Federation Member Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *FederationMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FederationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Permissions.GetFederationGroup(state.GroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Federation Member Failed", err.Error())
		return
	}

	// A member is identified within its group by organization (the API removes
	// by organizationId); the row id is checked too when the API returned one.
	var found *platformxe.FederationMember
	for i := range result.Members {
		m := &result.Members[i]
		if m.OrganizationID != state.OrganizationID.ValueString() {
			continue
		}
		if m.ID != "" && !state.ID.IsNull() && state.ID.ValueString() != "" && m.ID != state.ID.ValueString() {
			continue
		}
		found = m
		break
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if found.ID != "" {
		state.ID = types.StringValue(found.ID)
	}
	state.OrganizationID = types.StringValue(found.OrganizationID)
	state.Prefix = types.StringValue(found.Prefix)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *FederationMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Federation members are immutable — all attributes have RequiresReplace.
	// This method should never be called, but is required by the interface.
	resp.Diagnostics.AddError("Update Not Supported", "Federation members are immutable. Terraform will delete and recreate as needed.")
}

func (r *FederationMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FederationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Permissions.RemoveFederationMember(
		state.GroupID.ValueString(),
		state.OrganizationID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Delete Federation Member Failed", err.Error())
		return
	}
}
