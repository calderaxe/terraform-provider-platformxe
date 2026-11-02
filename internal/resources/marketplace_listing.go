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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

// platformxe_marketplace_listing — declarative state for a tenant
// custom-events marketplace listing (Phase 9C). PRO+ tenants publish
// a registered event so other PRO+ tenants can fork the schema.
//
// Maps to:
//   POST   /api/v1/events/custom/marketplace/publish     (create)
//   GET    /api/v1/events/custom/marketplace/[id]        (read)
//   DELETE /api/v1/events/custom/marketplace/[id]        (delete = unpublish)
//
// Update is a no-op (RequiresReplace on the schema-affecting fields).
// Republishing an unpublished listing is out of scope for the v1
// resource — operators rotate by deleting + recreating, or hit
// the API directly via `platformxe events:custom:marketplace:republish`
// (a separate runtime call, not infrastructure-shaped).

var _ resource.Resource = &MarketplaceListingResource{}
var _ resource.ResourceWithConfigure = &MarketplaceListingResource{}

type MarketplaceListingResource struct {
	client *platformxe.Client
}

type MarketplaceListingResourceModel struct {
	ID             types.String `tfsdk:"id"`
	RegistrationID types.String `tfsdk:"registration_id"`
	Title          types.String `tfsdk:"title"`
	Description    types.String `tfsdk:"description"`
	Tags           types.List   `tfsdk:"tags"`

	// Computed (server-assigned)
	Namespace           types.String `tfsdk:"namespace"`
	Name                types.String `tfsdk:"name"`
	Version             types.String `tfsdk:"version"`
	SourceCanonicalName types.String `tfsdk:"source_canonical_name"`
	Status              types.String `tfsdk:"status"`
	ForkCount           types.Int64  `tfsdk:"fork_count"`
	PublishedAt         types.String `tfsdk:"published_at"`
}

func NewMarketplaceListingResource() resource.Resource {
	return &MarketplaceListingResource{}
}

func (r *MarketplaceListingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_marketplace_listing"
}

func (r *MarketplaceListingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a tenant custom-events marketplace listing. PRO+ only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned listing id (mkl_…).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registration_id": schema.StringAttribute{
				Required:    true,
				Description: "Source registration id to publish (must be owned + status='published').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable label for the marketplace UI. Min 3 chars.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Longer description / use cases.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Search/filter tokens.",
			},
			"namespace": schema.StringAttribute{
				Computed:    true,
				Description: "Namespace cached at publish time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Event name cached at publish time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Version cached at publish time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_canonical_name": schema.StringAttribute{
				Computed:    true,
				Description: "TENANT_CUSTOM:<orgId>:<ns>.<name>@<version> at publish time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Lifecycle state: published / unpublished / archived.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fork_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of forks of this listing.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"published_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO-8601 timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MarketplaceListingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MarketplaceListingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MarketplaceListingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := platformxe.PublishMarketplaceInput{
		RegistrationID: plan.RegistrationID.ValueString(),
		Title:          plan.Title.ValueString(),
	}
	if !plan.Description.IsNull() {
		v := plan.Description.ValueString()
		input.Description = &v
	}
	if !plan.Tags.IsNull() {
		var tags []string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.Tags = tags
	}

	result, err := r.client.Events.Custom.Marketplace.Publish(input)
	if err != nil {
		resp.Diagnostics.AddError("Publish Marketplace Listing Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.Namespace = types.StringValue(result.Namespace)
	plan.Name = types.StringValue(result.Name)
	plan.Version = types.StringValue(result.Version)
	plan.SourceCanonicalName = types.StringValue(result.SourceCanonicalName)
	plan.Status = types.StringValue(result.Status)
	plan.ForkCount = types.Int64Value(int64(result.ForkCount))
	plan.PublishedAt = types.StringValue(result.PublishedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *MarketplaceListingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MarketplaceListingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := r.client.Events.Custom.Marketplace.Get(state.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Title = types.StringValue(detail.Title)
	if detail.Description != nil {
		state.Description = types.StringValue(*detail.Description)
	}
	state.Namespace = types.StringValue(detail.Namespace)
	state.Name = types.StringValue(detail.Name)
	state.Version = types.StringValue(detail.Version)
	state.SourceCanonicalName = types.StringValue(detail.SourceCanonicalName)
	state.Status = types.StringValue(detail.Status)
	state.ForkCount = types.Int64Value(int64(detail.ForkCount))
	state.PublishedAt = types.StringValue(detail.PublishedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *MarketplaceListingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All shape-affecting attributes (registration_id) have RequiresReplace.
	// Title/description/tags updates are not yet supported by the API
	// (would need a PATCH endpoint); accepted as state-only no-ops for now.
	var plan MarketplaceListingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *MarketplaceListingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MarketplaceListingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Events.Custom.Marketplace.Unpublish(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unpublish Marketplace Listing Failed", err.Error())
		return
	}
}
