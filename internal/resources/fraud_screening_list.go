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

var _ resource.Resource = &FraudScreeningListResource{}
var _ resource.ResourceWithConfigure = &FraudScreeningListResource{}

// FraudScreeningListResource manages a tenant-managed screening list.
//
// Two kinds are supported: `tenant_blocklist` and `tenant_allowlist`.
// The platform's admin-managed lists (sanctions / pep / adverse_media)
// are NOT manageable from Terraform — they are ingested by the platform's
// daily refresh cron.
//
// Entries are managed separately via the `platformxe_fraud_screening_list_entry`
// resource (or via the SDK's `appendEntries` helper for bulk loads).
type FraudScreeningListResource struct {
	client *platformxe.Client
}

type FraudScreeningListResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Source types.String `tfsdk:"source"`
	Name   types.String `tfsdk:"name"`
	Kind   types.String `tfsdk:"kind"`
}

func NewFraudScreeningListResource() resource.Resource {
	return &FraudScreeningListResource{}
}

func (r *FraudScreeningListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fraud_screening_list"
}

func (r *FraudScreeningListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a tenant-managed Fraud Detection screening list. " +
			"Use this for `tenant_blocklist` and `tenant_allowlist` lists; the " +
			"platform's global lists (sanctions / pep / adverse_media) are " +
			"admin-managed and cannot be created or updated from Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "List identifier (slst_<cuid>).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source": schema.StringAttribute{
				Required: true,
				Description: "Tenant-defined source identifier, unique within the " +
					"organisation. Immutable after create.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the list. Mutable.",
			},
			"kind": schema.StringAttribute{
				Required:    true,
				Description: "List kind: `tenant_blocklist` or `tenant_allowlist`. Immutable.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *FraudScreeningListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FraudScreeningListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FraudScreeningListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	row, err := r.client.Fraud.Lists.Create(
		plan.Source.ValueString(),
		plan.Name.ValueString(),
		plan.Kind.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create screening list", err.Error())
		return
	}
	applyFraudListResponseToState(row, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *FraudScreeningListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FraudScreeningListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	row, err := r.client.Fraud.Lists.Get(state.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	applyFraudListResponseToState(row, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *FraudScreeningListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FraudScreeningListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state FraudScreeningListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	row, err := r.client.Fraud.Lists.Update(state.ID.ValueString(), map[string]interface{}{
		"name": plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update screening list", err.Error())
		return
	}
	applyFraudListResponseToState(row, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *FraudScreeningListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FraudScreeningListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Fraud.Lists.Delete(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete screening list", err.Error())
		return
	}
}

func applyFraudListResponseToState(row map[string]interface{}, state *FraudScreeningListResourceModel) {
	if id, ok := row["id"].(string); ok {
		state.ID = types.StringValue(id)
	}
	if source, ok := row["source"].(string); ok {
		state.Source = types.StringValue(source)
	}
	if name, ok := row["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if kind, ok := row["kind"].(string); ok {
		state.Kind = types.StringValue(kind)
	}
}
