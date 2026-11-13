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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &ThreadChannelResource{}
var _ resource.ResourceWithConfigure = &ThreadChannelResource{}

type ThreadChannelResource struct {
	client *platformxe.Client
}

// -- Terraform models --------------------------------------------------------

type LifecycleAutoCloseModel struct {
	OnEntityStatus types.List `tfsdk:"on_entity_status"`
}

type LifecycleAutoArchiveModel struct {
	AfterClosedDays types.Int64 `tfsdk:"after_closed_days"`
}

type LifecycleSystemMessagesModel struct {
	OnThreadCreated types.String `tfsdk:"on_thread_created"`
	OnThreadClosed  types.String `tfsdk:"on_thread_closed"`
}

type LifecycleRulesModel struct {
	AutoClose      *LifecycleAutoCloseModel      `tfsdk:"auto_close"`
	AutoArchive    *LifecycleAutoArchiveModel    `tfsdk:"auto_archive"`
	SystemMessages *LifecycleSystemMessagesModel `tfsdk:"system_messages"`
}

type EscalationFlagReasonModel struct {
	Code     types.String `tfsdk:"code"`
	Label    types.String `tfsdk:"label"`
	Severity types.String `tfsdk:"severity"`
}

type EscalationAutoDetectionModel struct {
	Patterns   types.List   `tfsdk:"patterns"`
	FlagReason types.String `tfsdk:"flag_reason"`
	AutoFlag   types.Bool   `tfsdk:"auto_flag"`
}

type EscalationRuleActionModel struct {
	Type   types.String `tfsdk:"type"`
	Config types.String `tfsdk:"config"`
}

type EscalationRuleModel struct {
	ID              types.String                `tfsdk:"id"`
	Name            types.String                `tfsdk:"name"`
	Trigger         types.String                `tfsdk:"trigger"`
	Conditions      types.String                `tfsdk:"conditions"`
	Actions         []EscalationRuleActionModel `tfsdk:"actions"`
	Priority        types.Int64                 `tfsdk:"priority"`
	IsActive        types.Bool                  `tfsdk:"is_active"`
	CooldownMinutes types.Int64                 `tfsdk:"cooldown_minutes"`
}

type EscalationConfigModel struct {
	FlagReasons   []EscalationFlagReasonModel    `tfsdk:"flag_reasons"`
	AutoDetection []EscalationAutoDetectionModel `tfsdk:"auto_detection"`
	Rules         []EscalationRuleModel          `tfsdk:"rules"`
}

type ThreadChannelResourceModel struct {
	ID                types.String           `tfsdk:"id"`
	Slug              types.String           `tfsdk:"slug"`
	DisplayName       types.String           `tfsdk:"display_name"`
	EntityType        types.String           `tfsdk:"entity_type"`
	ParticipantRoles  types.List             `tfsdk:"participant_roles"`
	DefaultVisibility types.List             `tfsdk:"default_visibility"`
	LifecycleRules    *LifecycleRulesModel   `tfsdk:"lifecycle_rules"`
	EscalationConfig  *EscalationConfigModel `tfsdk:"escalation_config"`
	WebhookURL        types.String           `tfsdk:"webhook_url"`
	IsActive          types.Bool             `tfsdk:"is_active"`
}

func NewThreadsChannelResource() resource.Resource {
	return &ThreadChannelResource{}
}

func (r *ThreadChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_threads_channel"
}

func (r *ThreadChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe Contextual Messaging channel — defines entity type, participant roles, visibility rules, lifecycle automation, and escalation configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Channel identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Required:    true,
				Description: "Unique channel slug within the organization (e.g., 'booking', 'service-order', 'appointment').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable channel name (e.g., 'Booking Conversations').",
			},
			"entity_type": schema.StringAttribute{
				Required:    true,
				Description: "The business entity type this channel serves (e.g., 'BOOKING', 'SERVICE_ORDER', 'APPOINTMENT', 'CLAIM').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"participant_roles": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Allowed participant roles in this channel (e.g., ['GUEST', 'HOST', 'PLATFORM']).",
			},
			"default_visibility": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Default message visibility when sender doesn't specify (e.g., ['ALL']).",
			},
			"lifecycle_rules": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Lifecycle automation rules for threads in this channel.",
				Attributes: map[string]schema.Attribute{
					"auto_close": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Automatic thread closure settings.",
						Attributes: map[string]schema.Attribute{
							"on_entity_status": schema.ListAttribute{
								Required:    true,
								ElementType: types.StringType,
								Description: "Entity statuses that trigger automatic thread closure (e.g., ['CHECKED_OUT', 'CANCELLED']).",
							},
						},
					},
					"auto_archive": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Automatic thread archival settings.",
						Attributes: map[string]schema.Attribute{
							"after_closed_days": schema.Int64Attribute{
								Required:    true,
								Description: "Number of days after closure before automatic archival.",
							},
						},
					},
					"system_messages": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "System message templates for lifecycle events.",
						Attributes: map[string]schema.Attribute{
							"on_thread_created": schema.StringAttribute{
								Optional:    true,
								Description: "Message sent when a thread is created.",
							},
							"on_thread_closed": schema.StringAttribute{
								Optional:    true,
								Description: "Message sent when a thread is closed. Supports {closedReason} placeholder.",
							},
						},
					},
				},
			},
			"escalation_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Escalation configuration for flagging and auto-detection.",
				Attributes: map[string]schema.Attribute{
					"flag_reasons": schema.ListNestedAttribute{
						Optional:    true,
						Description: "Available flag reasons for participants.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"code": schema.StringAttribute{
									Required:    true,
									Description: "Flag reason code (e.g., SAFETY, DISPUTE).",
								},
								"label": schema.StringAttribute{
									Required:    true,
									Description: "Human-readable label.",
								},
								"severity": schema.StringAttribute{
									Required:    true,
									Description: "Severity level: LOW, MEDIUM, HIGH, or URGENT.",
								},
							},
						},
					},
					"auto_detection": schema.ListNestedAttribute{
						Optional:    true,
						Description: "Auto-detection patterns for flagging.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"patterns": schema.ListAttribute{
									Required:    true,
									ElementType: types.StringType,
									Description: "Text patterns to detect in messages.",
								},
								"flag_reason": schema.StringAttribute{
									Required:    true,
									Description: "Flag reason code to apply when patterns match.",
								},
								"auto_flag": schema.BoolAttribute{
									Required:    true,
									Description: "Whether to automatically flag matching messages.",
								},
							},
						},
					},
					"rules": schema.ListNestedAttribute{
						Optional:    true,
						Description: "Escalation rules using JSON Logic conditions.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required:    true,
									Description: "Rule identifier.",
								},
								"name": schema.StringAttribute{
									Required:    true,
									Description: "Rule name.",
								},
								"trigger": schema.StringAttribute{
									Required:    true,
									Description: "Trigger event (e.g., PARTICIPANT_FLAG).",
								},
								"conditions": schema.StringAttribute{
									Required:    true,
									Description: "JSON Logic condition expression.",
								},
								"actions": schema.ListNestedAttribute{
									Required:    true,
									Description: "Actions to execute when the rule fires.",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Required:    true,
												Description: "Action type (e.g., CREATE_ISSUE, NOTIFY_PARTICIPANTS, WEBHOOK).",
											},
											"config": schema.StringAttribute{
												Optional:    true,
												Description: "JSON-encoded action configuration.",
											},
										},
									},
								},
								"priority": schema.Int64Attribute{
									Required:    true,
									Description: "Rule evaluation priority.",
								},
								"is_active": schema.BoolAttribute{
									Required:    true,
									Description: "Whether the rule is active.",
								},
								"cooldown_minutes": schema.Int64Attribute{
									Optional:    true,
									Description: "Minimum minutes between consecutive rule firings.",
								},
							},
						},
					},
				},
			},
			"webhook_url": schema.StringAttribute{
				Optional:    true,
				Description: "Channel-level webhook URL for thread events.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the channel is active. Defaults to true.",
			},
		},
	}
}

func (r *ThreadChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// -- Conversion helpers: Terraform model <-> API map -------------------------

func lifecycleModelToMap(ctx context.Context, lc *LifecycleRulesModel) map[string]interface{} {
	if lc == nil {
		return nil
	}
	m := map[string]interface{}{}

	if lc.AutoClose != nil {
		var statuses []string
		lc.AutoClose.OnEntityStatus.ElementsAs(ctx, &statuses, false)
		m["autoClose"] = map[string]interface{}{
			"onEntityStatus": statuses,
		}
	}
	if lc.AutoArchive != nil {
		m["autoArchive"] = map[string]interface{}{
			"afterClosedDays": lc.AutoArchive.AfterClosedDays.ValueInt64(),
		}
	}
	if lc.SystemMessages != nil {
		sm := map[string]interface{}{}
		if !lc.SystemMessages.OnThreadCreated.IsNull() && !lc.SystemMessages.OnThreadCreated.IsUnknown() {
			sm["onThreadCreated"] = lc.SystemMessages.OnThreadCreated.ValueString()
		}
		if !lc.SystemMessages.OnThreadClosed.IsNull() && !lc.SystemMessages.OnThreadClosed.IsUnknown() {
			sm["onThreadClosed"] = lc.SystemMessages.OnThreadClosed.ValueString()
		}
		if len(sm) > 0 {
			m["systemMessages"] = sm
		}
	}

	return m
}

func escalationModelToMap(ctx context.Context, ec *EscalationConfigModel) map[string]interface{} {
	if ec == nil {
		return nil
	}
	m := map[string]interface{}{}

	if len(ec.FlagReasons) > 0 {
		reasons := make([]map[string]interface{}, 0, len(ec.FlagReasons))
		for _, fr := range ec.FlagReasons {
			reasons = append(reasons, map[string]interface{}{
				"code":     fr.Code.ValueString(),
				"label":    fr.Label.ValueString(),
				"severity": fr.Severity.ValueString(),
			})
		}
		m["flagReasons"] = reasons
	}

	if len(ec.AutoDetection) > 0 {
		detections := make([]map[string]interface{}, 0, len(ec.AutoDetection))
		for _, ad := range ec.AutoDetection {
			var patterns []string
			ad.Patterns.ElementsAs(ctx, &patterns, false)
			detections = append(detections, map[string]interface{}{
				"patterns":   patterns,
				"flagReason": ad.FlagReason.ValueString(),
				"autoFlag":   ad.AutoFlag.ValueBool(),
			})
		}
		m["autoDetection"] = detections
	}

	if len(ec.Rules) > 0 {
		rules := make([]map[string]interface{}, 0, len(ec.Rules))
		for _, rule := range ec.Rules {
			r := map[string]interface{}{
				"id":       rule.ID.ValueString(),
				"name":     rule.Name.ValueString(),
				"trigger":  rule.Trigger.ValueString(),
				"priority": rule.Priority.ValueInt64(),
				"isActive": rule.IsActive.ValueBool(),
			}
			// Parse JSON Logic conditions
			var cond interface{}
			if err := json.Unmarshal([]byte(rule.Conditions.ValueString()), &cond); err == nil {
				r["conditions"] = cond
			}
			if !rule.CooldownMinutes.IsNull() && !rule.CooldownMinutes.IsUnknown() {
				r["cooldownMinutes"] = rule.CooldownMinutes.ValueInt64()
			}
			// Actions
			actions := make([]map[string]interface{}, 0, len(rule.Actions))
			for _, a := range rule.Actions {
				action := map[string]interface{}{
					"type": a.Type.ValueString(),
				}
				if !a.Config.IsNull() && !a.Config.IsUnknown() {
					var cfg map[string]interface{}
					if err := json.Unmarshal([]byte(a.Config.ValueString()), &cfg); err == nil {
						action["config"] = cfg
					}
				}
				actions = append(actions, action)
			}
			r["actions"] = actions
			rules = append(rules, r)
		}
		m["rules"] = rules
	}

	return m
}

func lifecycleMapToModel(ctx context.Context, m map[string]interface{}) *LifecycleRulesModel {
	if m == nil || len(m) == 0 {
		return nil
	}
	model := &LifecycleRulesModel{}

	if ac, ok := m["autoClose"].(map[string]interface{}); ok {
		if statuses, ok := ac["onEntityStatus"].([]interface{}); ok {
			strs := make([]string, 0, len(statuses))
			for _, s := range statuses {
				if v, ok := s.(string); ok {
					strs = append(strs, v)
				}
			}
			list, _ := types.ListValueFrom(ctx, types.StringType, strs)
			model.AutoClose = &LifecycleAutoCloseModel{
				OnEntityStatus: list,
			}
		}
	}

	if aa, ok := m["autoArchive"].(map[string]interface{}); ok {
		if days, ok := aa["afterClosedDays"].(float64); ok {
			model.AutoArchive = &LifecycleAutoArchiveModel{
				AfterClosedDays: types.Int64Value(int64(days)),
			}
		}
	}

	if sm, ok := m["systemMessages"].(map[string]interface{}); ok {
		msgs := &LifecycleSystemMessagesModel{}
		if v, ok := sm["onThreadCreated"].(string); ok {
			msgs.OnThreadCreated = types.StringValue(v)
		} else {
			msgs.OnThreadCreated = types.StringNull()
		}
		if v, ok := sm["onThreadClosed"].(string); ok {
			msgs.OnThreadClosed = types.StringValue(v)
		} else {
			msgs.OnThreadClosed = types.StringNull()
		}
		model.SystemMessages = msgs
	}

	// Only return the model if at least one sub-block is set.
	if model.AutoClose == nil && model.AutoArchive == nil && model.SystemMessages == nil {
		return nil
	}
	return model
}

func escalationMapToModel(ctx context.Context, m map[string]interface{}) *EscalationConfigModel {
	if m == nil || len(m) == 0 {
		return nil
	}
	model := &EscalationConfigModel{}

	if reasons, ok := m["flagReasons"].([]interface{}); ok {
		model.FlagReasons = make([]EscalationFlagReasonModel, 0, len(reasons))
		for _, r := range reasons {
			rm, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			fr := EscalationFlagReasonModel{}
			if v, ok := rm["code"].(string); ok {
				fr.Code = types.StringValue(v)
			}
			if v, ok := rm["label"].(string); ok {
				fr.Label = types.StringValue(v)
			}
			if v, ok := rm["severity"].(string); ok {
				fr.Severity = types.StringValue(v)
			}
			model.FlagReasons = append(model.FlagReasons, fr)
		}
	}

	if detections, ok := m["autoDetection"].([]interface{}); ok {
		model.AutoDetection = make([]EscalationAutoDetectionModel, 0, len(detections))
		for _, d := range detections {
			dm, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			ad := EscalationAutoDetectionModel{}
			if patterns, ok := dm["patterns"].([]interface{}); ok {
				strs := make([]string, 0, len(patterns))
				for _, p := range patterns {
					if s, ok := p.(string); ok {
						strs = append(strs, s)
					}
				}
				list, _ := types.ListValueFrom(ctx, types.StringType, strs)
				ad.Patterns = list
			}
			if v, ok := dm["flagReason"].(string); ok {
				ad.FlagReason = types.StringValue(v)
			}
			if v, ok := dm["autoFlag"].(bool); ok {
				ad.AutoFlag = types.BoolValue(v)
			}
			model.AutoDetection = append(model.AutoDetection, ad)
		}
	}

	if rules, ok := m["rules"].([]interface{}); ok {
		model.Rules = make([]EscalationRuleModel, 0, len(rules))
		for _, r := range rules {
			rm, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			rule := EscalationRuleModel{}
			if v, ok := rm["id"].(string); ok {
				rule.ID = types.StringValue(v)
			}
			if v, ok := rm["name"].(string); ok {
				rule.Name = types.StringValue(v)
			}
			if v, ok := rm["trigger"].(string); ok {
				rule.Trigger = types.StringValue(v)
			}
			if v, ok := rm["priority"].(float64); ok {
				rule.Priority = types.Int64Value(int64(v))
			}
			if v, ok := rm["isActive"].(bool); ok {
				rule.IsActive = types.BoolValue(v)
			}
			if v, ok := rm["cooldownMinutes"].(float64); ok {
				rule.CooldownMinutes = types.Int64Value(int64(v))
			} else {
				rule.CooldownMinutes = types.Int64Null()
			}
			// Conditions — serialize back to JSON string
			if cond, ok := rm["conditions"]; ok {
				b, _ := json.Marshal(cond)
				rule.Conditions = types.StringValue(string(b))
			}
			// Actions
			if actions, ok := rm["actions"].([]interface{}); ok {
				rule.Actions = make([]EscalationRuleActionModel, 0, len(actions))
				for _, a := range actions {
					am, ok := a.(map[string]interface{})
					if !ok {
						continue
					}
					action := EscalationRuleActionModel{}
					if v, ok := am["type"].(string); ok {
						action.Type = types.StringValue(v)
					}
					if cfg, ok := am["config"].(map[string]interface{}); ok && len(cfg) > 0 {
						b, _ := json.Marshal(cfg)
						action.Config = types.StringValue(string(b))
					} else {
						action.Config = types.StringNull()
					}
					rule.Actions = append(rule.Actions, action)
				}
			}
			model.Rules = append(model.Rules, rule)
		}
	}

	// Only return the model if something is populated.
	if len(model.FlagReasons) == 0 && len(model.AutoDetection) == 0 && len(model.Rules) == 0 {
		return nil
	}
	return model
}

func (r *ThreadChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ThreadChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var roles []string
	resp.Diagnostics.Append(plan.ParticipantRoles.ElementsAs(ctx, &roles, false)...)
	var visibility []string
	resp.Diagnostics.Append(plan.DefaultVisibility.ElementsAs(ctx, &visibility, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := platformxe.CreateThreadChannelInput{
		Slug:              plan.Slug.ValueString(),
		DisplayName:       plan.DisplayName.ValueString(),
		EntityType:        plan.EntityType.ValueString(),
		ParticipantRoles:  roles,
		DefaultVisibility: visibility,
	}

	if lc := lifecycleModelToMap(ctx, plan.LifecycleRules); lc != nil {
		input.LifecycleRules = lc
	}
	if ec := escalationModelToMap(ctx, plan.EscalationConfig); ec != nil {
		input.EscalationConfig = ec
	}
	if !plan.WebhookURL.IsNull() && !plan.WebhookURL.IsUnknown() {
		input.WebhookURL = plan.WebhookURL.ValueString()
	}

	result, err := r.client.Threads.CreateChannel(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Thread Channel Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ThreadChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ThreadChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Threads.ListChannels()
	if err != nil {
		resp.Diagnostics.AddError("Read Thread Channels Failed", err.Error())
		return
	}

	var found *platformxe.ThreadChannel
	for i := range result.Channels {
		if result.Channels[i].ID == state.ID.ValueString() {
			found = &result.Channels[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.DisplayName = types.StringValue(found.DisplayName)
	state.IsActive = types.BoolValue(found.IsActive)
	state.Slug = types.StringValue(found.Slug)
	state.EntityType = types.StringValue(found.EntityType)

	if len(found.ParticipantRoles) > 0 {
		list, _ := types.ListValueFrom(ctx, types.StringType, found.ParticipantRoles)
		state.ParticipantRoles = list
	}
	if len(found.DefaultVisibility) > 0 {
		list, _ := types.ListValueFrom(ctx, types.StringType, found.DefaultVisibility)
		state.DefaultVisibility = list
	}
	if found.WebhookURL != "" {
		state.WebhookURL = types.StringValue(found.WebhookURL)
	}

	state.LifecycleRules = lifecycleMapToModel(ctx, found.LifecycleRules)
	state.EscalationConfig = escalationMapToModel(ctx, found.EscalationConfig)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ThreadChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ThreadChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := platformxe.UpdateThreadChannelInput{}

	displayName := plan.DisplayName.ValueString()
	input.DisplayName = &displayName

	if !plan.IsActive.IsNull() {
		isActive := plan.IsActive.ValueBool()
		input.IsActive = &isActive
	}

	if lc := lifecycleModelToMap(ctx, plan.LifecycleRules); lc != nil {
		input.LifecycleRules = lc
	}
	if ec := escalationModelToMap(ctx, plan.EscalationConfig); ec != nil {
		input.EscalationConfig = ec
	}

	_, err := r.client.Threads.UpdateChannel(plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Update Thread Channel Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ThreadChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Channels are deactivated, not deleted — set isActive to false
	var state ThreadChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isActive := false
	_, err := r.client.Threads.UpdateChannel(state.ID.ValueString(), platformxe.UpdateThreadChannelInput{
		IsActive: &isActive,
	})
	if err != nil {
		resp.Diagnostics.AddError("Deactivate Thread Channel Failed", err.Error())
		return
	}
}
