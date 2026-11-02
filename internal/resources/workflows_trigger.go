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

var _ resource.Resource = &WorkflowResource{}
var _ resource.ResourceWithConfigure = &WorkflowResource{}

type WorkflowResource struct {
	client *platformxe.Client
}

// -- Terraform models --------------------------------------------------------

type WorkflowTriggerConfigModel struct {
	TriggerType    types.String `tfsdk:"trigger_type"`
	EventType      types.String `tfsdk:"event_type"`
	CronExpression types.String `tfsdk:"cron_expression"`
	EntityType     types.String `tfsdk:"entity_type"`
	Conditions     types.String `tfsdk:"conditions"`
}

type WorkflowActionModel struct {
	Type   types.String `tfsdk:"type"`
	Config types.String `tfsdk:"config"`
}

type WorkflowResourceModel struct {
	ID               types.String                `tfsdk:"id"`
	Name             types.String                `tfsdk:"name"`
	Description      types.String                `tfsdk:"description"`
	Code             types.String                `tfsdk:"code"`
	TriggerConfig    *WorkflowTriggerConfigModel `tfsdk:"trigger_config"`
	Actions          []WorkflowActionModel       `tfsdk:"actions"`
	IsActive         types.Bool                  `tfsdk:"is_active"`
	Priority         types.Int64                 `tfsdk:"priority"`
	CooldownMinutes  types.Int64                 `tfsdk:"cooldown_minutes"`
	MaxExecutionsDay types.Int64                 `tfsdk:"max_executions_day"`
	SourceApp        types.String                `tfsdk:"source_app"`
}

func NewWorkflowsTriggerResource() resource.Resource {
	return &WorkflowResource{}
}

func (r *WorkflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows_trigger"
}

func (r *WorkflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe workflow automation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Workflow identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Workflow name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Workflow description.",
			},
			"code": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Unique workflow code. Auto-generated if not provided.",
			},
			"trigger_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Trigger configuration defining when the workflow fires.",
				Attributes: map[string]schema.Attribute{
					"trigger_type": schema.StringAttribute{
						Required:    true,
						Description: "Trigger type: EVENT, CRON, or MANUAL.",
					},
					"event_type": schema.StringAttribute{
						Optional:    true,
						Description: "Event type to match (required when trigger_type is EVENT).",
					},
					"cron_expression": schema.StringAttribute{
						Optional:    true,
						Description: "Cron expression for scheduled triggers (required when trigger_type is CRON).",
					},
					"entity_type": schema.StringAttribute{
						Optional:    true,
						Description: "Entity type filter for the trigger.",
					},
					"conditions": schema.StringAttribute{
						Optional:    true,
						Description: "JSON-encoded condition object for trigger matching.",
					},
				},
			},
			"actions": schema.ListNestedAttribute{
				Required:    true,
				Description: "Ordered list of actions to execute when the workflow triggers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Action type (e.g., send_email, webhook, delay).",
						},
						"config": schema.StringAttribute{
							Optional:    true,
							Description: "JSON-encoded action configuration.",
						},
					},
				},
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the workflow is active. Defaults to true.",
			},
			"priority": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Evaluation priority. Higher values take precedence.",
			},
			"cooldown_minutes": schema.Int64Attribute{
				Optional:    true,
				Description: "Minimum minutes between consecutive executions.",
			},
			"max_executions_day": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum executions per day.",
			},
			"source_app": schema.StringAttribute{
				Optional:    true,
				Description: "Source application identifier.",
			},
		},
	}
}

func (r *WorkflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// buildWorkflowInput converts the Terraform model into the map the SDK expects.
func buildWorkflowInput(plan *WorkflowResourceModel) map[string]interface{} {
	input := map[string]interface{}{
		"name":     plan.Name.ValueString(),
		"isActive": plan.IsActive.ValueBool(),
		"priority": plan.Priority.ValueInt64(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		input["description"] = plan.Description.ValueString()
	}
	if !plan.Code.IsNull() && !plan.Code.IsUnknown() {
		input["code"] = plan.Code.ValueString()
	}
	if !plan.SourceApp.IsNull() && !plan.SourceApp.IsUnknown() {
		input["sourceApp"] = plan.SourceApp.ValueString()
	}
	if !plan.CooldownMinutes.IsNull() && !plan.CooldownMinutes.IsUnknown() {
		v := int(plan.CooldownMinutes.ValueInt64())
		input["cooldownMinutes"] = v
	}
	if !plan.MaxExecutionsDay.IsNull() && !plan.MaxExecutionsDay.IsUnknown() {
		v := int(plan.MaxExecutionsDay.ValueInt64())
		input["maxExecutionsDay"] = v
	}

	// Trigger config
	if tc := plan.TriggerConfig; tc != nil {
		input["triggerType"] = tc.TriggerType.ValueString()
		if !tc.EventType.IsNull() && !tc.EventType.IsUnknown() {
			input["eventType"] = tc.EventType.ValueString()
		}
		if !tc.CronExpression.IsNull() && !tc.CronExpression.IsUnknown() {
			input["cronExpression"] = tc.CronExpression.ValueString()
		}
		if !tc.EntityType.IsNull() && !tc.EntityType.IsUnknown() {
			input["entityType"] = tc.EntityType.ValueString()
		}
		if !tc.Conditions.IsNull() && !tc.Conditions.IsUnknown() {
			var cond map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Conditions.ValueString()), &cond); err == nil {
				input["conditions"] = cond
			}
		}
	}

	// Actions
	actions := make([]map[string]interface{}, 0, len(plan.Actions))
	for _, a := range plan.Actions {
		action := map[string]interface{}{
			"type": a.Type.ValueString(),
		}
		if !a.Config.IsNull() && !a.Config.IsUnknown() {
			var cfg map[string]interface{}
			if err := json.Unmarshal([]byte(a.Config.ValueString()), &cfg); err == nil {
				// Merge config fields into the action map (API expects flat action objects)
				for k, v := range cfg {
					action[k] = v
				}
			}
		}
		actions = append(actions, action)
	}
	input["actions"] = actions

	return input
}

// mapWorkflowToState populates a Terraform model from a typed SDK WorkflowTrigger.
func mapWorkflowToState(wf *platformxe.WorkflowTrigger, state *WorkflowResourceModel) {
	state.ID = types.StringValue(wf.ID)
	state.Name = types.StringValue(wf.Name)
	state.IsActive = types.BoolValue(wf.IsActive)
	state.Priority = types.Int64Value(int64(wf.Priority))

	if wf.Description != "" {
		state.Description = types.StringValue(wf.Description)
	}
	if wf.Code != "" {
		state.Code = types.StringValue(wf.Code)
	}
	if wf.SourceApp != "" {
		state.SourceApp = types.StringValue(wf.SourceApp)
	}
	if wf.CooldownMinutes != nil {
		state.CooldownMinutes = types.Int64Value(int64(*wf.CooldownMinutes))
	}
	if wf.MaxExecutionsDay != nil {
		state.MaxExecutionsDay = types.Int64Value(int64(*wf.MaxExecutionsDay))
	}

	// Trigger config
	state.TriggerConfig = &WorkflowTriggerConfigModel{
		TriggerType: types.StringValue(wf.TriggerType),
	}
	if wf.EventType != "" {
		state.TriggerConfig.EventType = types.StringValue(wf.EventType)
	} else {
		state.TriggerConfig.EventType = types.StringNull()
	}
	if wf.CronExpression != "" {
		state.TriggerConfig.CronExpression = types.StringValue(wf.CronExpression)
	} else {
		state.TriggerConfig.CronExpression = types.StringNull()
	}
	if wf.EntityType != "" {
		state.TriggerConfig.EntityType = types.StringValue(wf.EntityType)
	} else {
		state.TriggerConfig.EntityType = types.StringNull()
	}
	if wf.Conditions != nil && len(wf.Conditions) > 0 {
		b, _ := json.Marshal(wf.Conditions)
		state.TriggerConfig.Conditions = types.StringValue(string(b))
	} else {
		state.TriggerConfig.Conditions = types.StringNull()
	}

	// Actions
	state.Actions = make([]WorkflowActionModel, 0, len(wf.Actions))
	for _, a := range wf.Actions {
		actionType, _ := a["type"].(string)
		// Build config from all fields except "type"
		cfg := make(map[string]interface{})
		for k, v := range a {
			if k != "type" {
				cfg[k] = v
			}
		}
		am := WorkflowActionModel{
			Type: types.StringValue(actionType),
		}
		if len(cfg) > 0 {
			b, _ := json.Marshal(cfg)
			am.Config = types.StringValue(string(b))
		} else {
			am.Config = types.StringNull()
		}
		state.Actions = append(state.Actions, am)
	}
}

func (r *WorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildWorkflowInput(&plan)

	result, err := r.client.Workflows.Create(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Workflow Failed", err.Error())
		return
	}

	mapWorkflowToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *WorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Workflows.Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Workflow Failed", err.Error())
		return
	}

	mapWorkflowToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *WorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildWorkflowInput(&plan)

	result, err := r.client.Workflows.Update(plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Update Workflow Failed", err.Error())
		return
	}

	mapWorkflowToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *WorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Workflows.Delete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Workflow Failed", err.Error())
		return
	}
}
