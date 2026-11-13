// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package resources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

// platformxe_custom_event resource — declarative tenant custom events.
//
// Maps to /api/v1/events/custom (POST/GET/DELETE) on PlatformXe. Emit,
// dry-run, and health are runtime-only and have no Terraform surface.

var _ resource.Resource = &CustomEventResource{}
var _ resource.ResourceWithConfigure = &CustomEventResource{}

type CustomEventResource struct {
	client *platformxe.Client
}

type CustomEventResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	Version        types.String `tfsdk:"version"`
	Status         types.String `tfsdk:"status"`
	Description    types.String `tfsdk:"description"`
	PayloadSchema  types.String `tfsdk:"payload_schema"`
	PayloadExample types.String `tfsdk:"payload_example"`

	// Computed (server-assigned)
	CanonicalName types.String `tfsdk:"canonical_name"`
	SchemaHash    types.String `tfsdk:"schema_hash"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewCustomEventResource() resource.Resource {
	return &CustomEventResource{}
}

func (r *CustomEventResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_event"
}

func (r *CustomEventResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a tenant-defined custom event registered with the PlatformXe events engine. " +
			"Schemas are immutable per (namespace, name, version) — bump the version to evolve a shape.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned registration id (cer_…).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Required:    true,
				Description: "Namespace this event belongs to. Must be on the org's allowlist (seeded with your slug at onboarding).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Event name within the namespace (e.g. \"property.favorited\").",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "Semantic version (MAJOR.MINOR.PATCH). Schemas are immutable per version.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Lifecycle status. One of \"draft\", \"published\", or \"archived\". Defaults to \"draft\".",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Human-readable description (also surfaced on the marketplace and the portal).",
			},
			"payload_schema": schema.StringAttribute{
				Required: true,
				Description: "JSON Schema 2020-12 document as a string (use jsonencode({...})). " +
					"OpenAPI 3.1 component schemas accepted as a strict superset.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"payload_example": schema.StringAttribute{
				Optional:    true,
				Description: "Optional example payload as a JSON string. Validated against payload_schema at register time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"canonical_name": schema.StringAttribute{
				Computed:    true,
				Description: "Server-derived bus-stable identifier (TENANT_CUSTOM:<orgId>:<ns>.<name>@<version>).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schema_hash": schema.StringAttribute{
				Computed:    true,
				Description: "SHA-256 of the canonical-JSON form of payload_schema.",
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
		},
	}
}

func (r *CustomEventResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CustomEventResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CustomEventResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var schemaObj map[string]interface{}
	if err := json.Unmarshal([]byte(plan.PayloadSchema.ValueString()), &schemaObj); err != nil {
		resp.Diagnostics.AddError("payload_schema is not valid JSON", err.Error())
		return
	}

	input := platformxe.RegisterCustomEventInput{
		Namespace:     plan.Namespace.ValueString(),
		Name:          plan.Name.ValueString(),
		Version:       plan.Version.ValueString(),
		PayloadSchema: schemaObj,
	}
	if !plan.Status.IsNull() && plan.Status.ValueString() != "" {
		s := platformxe.CustomEventStatus(plan.Status.ValueString())
		input.Status = &s
	}
	if !plan.Description.IsNull() {
		v := plan.Description.ValueString()
		input.Description = &v
	}
	if !plan.PayloadExample.IsNull() && plan.PayloadExample.ValueString() != "" {
		var exampleObj map[string]interface{}
		if err := json.Unmarshal([]byte(plan.PayloadExample.ValueString()), &exampleObj); err != nil {
			resp.Diagnostics.AddError("payload_example is not valid JSON", err.Error())
			return
		}
		input.PayloadExample = exampleObj
	}

	result, err := r.client.Events.Custom.Register(input)
	if err != nil {
		resp.Diagnostics.AddError("Register Custom Event Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.Status = types.StringValue(result.Status)
	plan.CanonicalName = types.StringValue(result.CanonicalName)
	plan.SchemaHash = types.StringValue(result.SchemaHash)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CustomEventResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomEventResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := r.client.Events.Custom.Get(state.ID.ValueString())
	if err != nil {
		// 404 → drop from state
		resp.State.RemoveResource(ctx)
		return
	}

	state.Namespace = types.StringValue(detail.Namespace)
	state.Name = types.StringValue(detail.Name)
	state.Version = types.StringValue(detail.Version)
	state.Status = types.StringValue(detail.Status)
	if detail.Description != nil {
		state.Description = types.StringValue(*detail.Description)
	}
	state.CanonicalName = types.StringValue(detail.CanonicalName)
	state.CreatedAt = types.StringValue(detail.CreatedAt)
	state.UpdatedAt = types.StringValue(detail.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CustomEventResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All shape-affecting attributes (namespace, name, version, payload_schema)
	// have RequiresReplace plan modifiers, so an Update only sees changes to
	// description / status. We accept those as no-ops at the API today and
	// will wire a PATCH /events/custom/{id} when the engine supports it.
	var plan CustomEventResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CustomEventResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomEventResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Events.Custom.Archive(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Archive Custom Event Failed", err.Error())
		return
	}
}
