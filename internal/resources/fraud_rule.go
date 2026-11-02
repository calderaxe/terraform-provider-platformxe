// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &FraudRuleResource{}
var _ resource.ResourceWithConfigure = &FraudRuleResource{}

// FraudRuleResource manages a tenant-authored Fraud Detection rule.
//
// New rules are created in `draft` status. To promote them through
// shadow → published, set `status` accordingly. Deleting the resource
// archives the rule (terminal — fraud rules are append-only on the
// platform side).
type FraudRuleResource struct {
	client *platformxe.Client
}

type FraudRuleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Status          types.String `tfsdk:"status"`
	Weight          types.Int64  `tfsdk:"weight"`
	VerdictOverride types.String `tfsdk:"verdict_override"`
	Description     types.String `tfsdk:"description"`
	Version         types.Int64  `tfsdk:"version"`
	// JSON-encoded payloads for the AST + appliesTo + windows. Terraform's
	// dynamic schema doesn't model the rule DSL natively; encode as string.
	AppliesToJSON types.String `tfsdk:"applies_to_json"`
	ConditionJSON types.String `tfsdk:"condition_json"`
	WindowsJSON   types.String `tfsdk:"windows_json"`
}

func NewFraudRuleResource() resource.Resource {
	return &FraudRuleResource{}
}

func (r *FraudRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fraud_rule"
}

func (r *FraudRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe Fraud Detection rule. Rules are tenant-authored " +
			"predicates evaluated by the Detection Engine on every decide call. " +
			"Lifecycle: draft → shadow → published → archived. Deleting the " +
			"resource archives the rule (terminal).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Rule identifier (frl_<cuid>).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name. ≤ 200 characters.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Lifecycle status: draft | shadow | published | archived. Defaults to 'draft'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"weight": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Description: "Score contribution when the rule triggers. Defaults to 10.",
			},
			"verdict_override": schema.StringAttribute{
				Optional: true,
				Description: "Force a specific verdict on trigger regardless of score band " +
					"(allow | review | step_up | block). Used for sanctions hits.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Free-form description of what the rule detects.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Server-assigned version. Increments on every update.",
			},
			"applies_to_json": schema.StringAttribute{
				Required: true,
				Description: "JSON-encoded `appliesTo` object: " +
					"`{\"actions\": [\"...\"], \"resourceKinds\": [\"...\"]}`.",
			},
			"condition_json": schema.StringAttribute{
				Required: true,
				Description: "JSON-encoded rule condition AST. Supports the 13-operator " +
					"DSL plus all/any/not combinators and $count.<window> references.",
			},
			"windows_json": schema.StringAttribute{
				Optional: true,
				Description: "JSON-encoded array of velocity-counter window definitions. " +
					"Empty / null when the rule does not reference $count.<window>.",
			},
		},
	}
}

func (r *FraudRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FraudRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FraudRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildFraudRulePayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule payload", err.Error())
		return
	}

	row, err := r.client.Fraud.Rules.Create(payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create fraud rule", err.Error())
		return
	}

	applyFraudRuleResponseToState(row, &plan)

	// Optional: if the user requested a non-draft status, transition.
	desiredStatus := plan.Status.ValueString()
	if desiredStatus != "" && desiredStatus != "draft" {
		if err := transitionFraudRule(r.client, plan.ID.ValueString(), desiredStatus); err != nil {
			resp.Diagnostics.AddError("Failed to transition rule status", err.Error())
			return
		}
		// Re-read so state reflects the new status.
		row, err := r.client.Fraud.Rules.Get(plan.ID.ValueString())
		if err == nil {
			applyFraudRuleResponseToState(row, &plan)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *FraudRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FraudRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	row, err := r.client.Fraud.Rules.Get(state.ID.ValueString())
	if err != nil {
		// If the rule was archived externally, drop from state.
		resp.State.RemoveResource(ctx)
		return
	}
	applyFraudRuleResponseToState(row, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *FraudRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FraudRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state FraudRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildFraudRulePayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule payload", err.Error())
		return
	}

	row, err := r.client.Fraud.Rules.Update(state.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update fraud rule", err.Error())
		return
	}
	applyFraudRuleResponseToState(row, &plan)

	// If status changed, transition.
	if plan.Status.ValueString() != state.Status.ValueString() && !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		if err := transitionFraudRule(r.client, state.ID.ValueString(), plan.Status.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to transition rule status", err.Error())
			return
		}
		row, err := r.client.Fraud.Rules.Get(state.ID.ValueString())
		if err == nil {
			applyFraudRuleResponseToState(row, &plan)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *FraudRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FraudRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Fraud.Rules.Archive(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to archive fraud rule", err.Error())
		return
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func buildFraudRulePayload(plan FraudRuleResourceModel) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		payload["weight"] = plan.Weight.ValueInt64()
	}
	if !plan.VerdictOverride.IsNull() && !plan.VerdictOverride.IsUnknown() && plan.VerdictOverride.ValueString() != "" {
		payload["verdictOverride"] = plan.VerdictOverride.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	var appliesTo map[string]interface{}
	if err := json.Unmarshal([]byte(plan.AppliesToJSON.ValueString()), &appliesTo); err != nil {
		return nil, fmt.Errorf("applies_to_json is not valid JSON: %w", err)
	}
	payload["appliesTo"] = appliesTo

	var condition interface{}
	if err := json.Unmarshal([]byte(plan.ConditionJSON.ValueString()), &condition); err != nil {
		return nil, fmt.Errorf("condition_json is not valid JSON: %w", err)
	}
	payload["condition"] = condition

	if !plan.WindowsJSON.IsNull() && !plan.WindowsJSON.IsUnknown() && plan.WindowsJSON.ValueString() != "" {
		var windows []interface{}
		if err := json.Unmarshal([]byte(plan.WindowsJSON.ValueString()), &windows); err != nil {
			return nil, fmt.Errorf("windows_json is not valid JSON: %w", err)
		}
		payload["windows"] = windows
	}

	return payload, nil
}

func applyFraudRuleResponseToState(row map[string]interface{}, state *FraudRuleResourceModel) {
	if id, ok := row["id"].(string); ok {
		state.ID = types.StringValue(id)
	}
	if name, ok := row["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if status, ok := row["status"].(string); ok {
		state.Status = types.StringValue(status)
	}
	if weight, ok := row["weight"].(float64); ok {
		state.Weight = types.Int64Value(int64(weight))
	}
	if version, ok := row["version"].(float64); ok {
		state.Version = types.Int64Value(int64(version))
	}
	if v, ok := row["verdictOverride"].(string); ok && v != "" {
		state.VerdictOverride = types.StringValue(v)
	}
	if d, ok := row["description"].(string); ok && d != "" {
		state.Description = types.StringValue(d)
	}
}

// transitionFraudRule moves a rule through the lifecycle by calling the
// Publish or Archive endpoints. Shadow promotion is handled by the SDK
// helper when shipped; for now this maps "shadow" / "published" /
// "archived" to the existing client methods.
func transitionFraudRule(client *platformxe.Client, ruleID, target string) error {
	switch target {
	case "published":
		_, err := client.Fraud.Rules.Publish(ruleID)
		return err
	case "archived":
		_, err := client.Fraud.Rules.Archive(ruleID)
		return err
	case "shadow", "draft":
		// The Go SDK does not yet expose a direct transition helper for
		// these states. The PATCH helper above leaves the rule in its
		// current status; the user can drive shadow/draft via the API
		// directly. This is intentionally a no-op until the helper is
		// added in a future release.
		return nil
	default:
		return fmt.Errorf("unsupported status transition target: %s", target)
	}
}
