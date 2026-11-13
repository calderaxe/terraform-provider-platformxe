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

var _ resource.Resource = &FederationGroupResource{}
var _ resource.ResourceWithConfigure = &FederationGroupResource{}

type FederationGroupResource struct {
	client *platformxe.Client
}

type FederationGroupResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewPermissionsFederationGroupResource() resource.Resource {
	return &FederationGroupResource{}
}

func (r *FederationGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_federation_group"
}

func (r *FederationGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe federation group for multi-app permission orchestration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Federation group identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Federation group name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *FederationGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FederationGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FederationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Permissions.CreateFederationGroup(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Federation Group Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *FederationGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FederationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Permissions.GetFederationGroup(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Federation Group Failed", err.Error())
		return
	}

	state.Name = types.StringValue(result.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *FederationGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Federation groups are name-only and immutable — name has RequiresReplace.
	// This method should never be called, but is required by the interface.
	resp.Diagnostics.AddError("Update Not Supported", "Federation groups are immutable. Terraform will delete and recreate as needed.")
}

func (r *FederationGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FederationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Permissions.DeleteFederationGroup(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Federation Group Failed", err.Error())
		return
	}
}
