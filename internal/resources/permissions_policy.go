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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &PolicyResource{}
var _ resource.ResourceWithConfigure = &PolicyResource{}

type PolicyResource struct {
	client *platformxe.Client
}

// -- Terraform models --------------------------------------------------------

type PolicyConditionRuleModel struct {
	Attribute types.String `tfsdk:"attribute"`
	Operator  types.String `tfsdk:"operator"`
	Value     types.String `tfsdk:"value"`
}

type PolicyConditionModel struct {
	Combinator types.String               `tfsdk:"combinator"`
	Rules      []PolicyConditionRuleModel `tfsdk:"rules"`
}

type PolicyResourceModel struct {
	ID          types.String          `tfsdk:"id"`
	Path        types.String          `tfsdk:"path"`
	Action      types.String          `tfsdk:"action"`
	Condition   *PolicyConditionModel `tfsdk:"condition"`
	Effect      types.String          `tfsdk:"effect"`
	Priority    types.Int64           `tfsdk:"priority"`
	Description types.String          `tfsdk:"description"`
	IsActive    types.Bool            `tfsdk:"is_active"`
}

func NewPermissionsPolicyResource() resource.Resource {
	return &PolicyResource{}
}

func (r *PolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_policy"
}

func (r *PolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe resource policy (ABAC).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Policy identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Resource path pattern the policy applies to.",
			},
			"action": schema.StringAttribute{
				Required:    true,
				Description: "Action the policy governs (e.g. read, write, delete).",
			},
			"condition": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "ABAC condition expression with combinator logic.",
				Attributes: map[string]schema.Attribute{
					"combinator": schema.StringAttribute{
						Required:    true,
						Description: "Logic combinator: all, any, or not.",
					},
					"rules": schema.ListNestedAttribute{
						Required:    true,
						Description: "List of condition rules to evaluate.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"attribute": schema.StringAttribute{
									Required:    true,
									Description: "Attribute path to evaluate (e.g., user.region, resource.owner).",
								},
								"operator": schema.StringAttribute{
									Required:    true,
									Description: "Comparison operator: eq, neq, gt, gte, lt, lte, in, not_in, contains, starts_with, ends_with, exists, not_exists.",
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "Value to compare against. For array operators (in, not_in), use JSON-encoded array.",
								},
							},
						},
					},
				},
			},
			"effect": schema.StringAttribute{
				Required:    true,
				Description: "Policy effect: ALLOW or DENY.",
			},
			"priority": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Policy evaluation priority. Higher values take precedence.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Policy description.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the policy is active. Defaults to true.",
			},
		},
	}
}

func (r *PolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// conditionModelToMap converts the typed condition block to the map the API expects.
func conditionModelToMap(cond *PolicyConditionModel) map[string]interface{} {
	if cond == nil {
		return nil
	}
	rules := make([]interface{}, 0, len(cond.Rules))
	for _, r := range cond.Rules {
		rule := map[string]interface{}{
			"attribute": r.Attribute.ValueString(),
			"operator":  r.Operator.ValueString(),
		}
		// Try to parse value as JSON for array operators; fall back to string.
		var parsed interface{}
		if err := json.Unmarshal([]byte(r.Value.ValueString()), &parsed); err == nil {
			rule["value"] = parsed
		} else {
			rule["value"] = r.Value.ValueString()
		}
		rules = append(rules, rule)
	}
	return map[string]interface{}{
		cond.Combinator.ValueString(): rules,
	}
}

// conditionMapToModel converts the API condition map to the typed Terraform model.
func conditionMapToModel(cond map[string]interface{}) *PolicyConditionModel {
	if cond == nil || len(cond) == 0 {
		return nil
	}
	for combinator, v := range cond {
		rules, ok := v.([]interface{})
		if !ok {
			continue
		}
		model := &PolicyConditionModel{
			Combinator: types.StringValue(combinator),
			Rules:      make([]PolicyConditionRuleModel, 0, len(rules)),
		}
		for _, r := range rules {
			rm, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			rule := PolicyConditionRuleModel{}
			if attr, ok := rm["attribute"].(string); ok {
				rule.Attribute = types.StringValue(attr)
			}
			if op, ok := rm["operator"].(string); ok {
				rule.Operator = types.StringValue(op)
			}
			// Serialize value to string — arrays become JSON, scalars stay as-is.
			if val, ok := rm["value"]; ok {
				switch v := val.(type) {
				case string:
					rule.Value = types.StringValue(v)
				default:
					b, _ := json.Marshal(v)
					rule.Value = types.StringValue(string(b))
				}
			}
			model.Rules = append(model.Rules, rule)
		}
		return model // Only one combinator key expected at top level.
	}
	return nil
}

func buildPolicyInput(plan *PolicyResourceModel) map[string]interface{} {
	input := map[string]interface{}{
		"path":     plan.Path.ValueString(),
		"action":   plan.Action.ValueString(),
		"effect":   plan.Effect.ValueString(),
		"priority": plan.Priority.ValueInt64(),
		"isActive": plan.IsActive.ValueBool(),
	}
	if plan.Condition != nil {
		input["condition"] = conditionModelToMap(plan.Condition)
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		input["description"] = plan.Description.ValueString()
	}
	return input
}

func mapPolicyToState(p *platformxe.ResourcePolicy, state *PolicyResourceModel) {
	state.ID = types.StringValue(p.ID)
	state.Path = types.StringValue(p.Path)
	state.Action = types.StringValue(p.Action)
	state.Effect = types.StringValue(p.Effect)
	state.Priority = types.Int64Value(int64(p.Priority))
	state.IsActive = types.BoolValue(p.IsActive)

	if p.Description != "" {
		state.Description = types.StringValue(p.Description)
	} else {
		state.Description = types.StringNull()
	}

	state.Condition = conditionMapToModel(p.Condition)
}

func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildPolicyInput(&plan)

	result, err := r.client.Permissions.CreatePolicy(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Policy Failed", err.Error())
		return
	}

	mapPolicyToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Permissions.ListPolicies()
	if err != nil {
		resp.Diagnostics.AddError("Read Policy Failed", err.Error())
		return
	}

	var found *platformxe.ResourcePolicy
	for i := range result.Policies {
		if result.Policies[i].ID == state.ID.ValueString() {
			found = &result.Policies[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapPolicyToState(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildPolicyInput(&plan)

	result, err := r.client.Permissions.UpdatePolicy(plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Update Policy Failed", err.Error())
		return
	}

	mapPolicyToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Permissions.DeletePolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Policy Failed", err.Error())
		return
	}
}
