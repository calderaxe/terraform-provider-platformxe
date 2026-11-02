// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package datasources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ datasource.DataSource = &ThreadsEscalationConfigDataSource{}
var _ datasource.DataSourceWithConfigure = &ThreadsEscalationConfigDataSource{}

// ThreadsEscalationConfigDataSource reads the escalation configuration for a thread channel.
type ThreadsEscalationConfigDataSource struct {
	client *platformxe.Client
}

type EscalationFlagReasonItem struct {
	Code     types.String `tfsdk:"code"`
	Label    types.String `tfsdk:"label"`
	Severity types.String `tfsdk:"severity"`
}

type EscalationAutoDetectionItem struct {
	Patterns   types.List   `tfsdk:"patterns"`
	FlagReason types.String `tfsdk:"flag_reason"`
	AutoFlag   types.Bool   `tfsdk:"auto_flag"`
}

type EscalationRuleActionItem struct {
	Type   types.String `tfsdk:"type"`
	Config types.String `tfsdk:"config"`
}

type EscalationRuleItem struct {
	ID              types.String               `tfsdk:"id"`
	Name            types.String               `tfsdk:"name"`
	Trigger         types.String               `tfsdk:"trigger"`
	Conditions      types.String               `tfsdk:"conditions"`
	Actions         []EscalationRuleActionItem `tfsdk:"actions"`
	Priority        types.Int64                `tfsdk:"priority"`
	IsActive        types.Bool                 `tfsdk:"is_active"`
	CooldownMinutes types.Int64                `tfsdk:"cooldown_minutes"`
}

type ThreadsEscalationConfigDataSourceModel struct {
	ChannelID     types.String                  `tfsdk:"channel_id"`
	FlagReasons   []EscalationFlagReasonItem    `tfsdk:"flag_reasons"`
	AutoDetection []EscalationAutoDetectionItem `tfsdk:"auto_detection"`
	Rules         []EscalationRuleItem          `tfsdk:"rules"`
}

func NewThreadsEscalationConfigDataSource() datasource.DataSource {
	return &ThreadsEscalationConfigDataSource{}
}

func (d *ThreadsEscalationConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_threads_escalation_config"
}

func (d *ThreadsEscalationConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the escalation configuration for a PlatformXe Contextual Messaging channel. Use this to inspect flag reasons, auto-detection patterns, and escalation rules.",
		Attributes: map[string]schema.Attribute{
			"channel_id": schema.StringAttribute{
				Required:    true,
				Description: "The channel ID to read escalation config from.",
			},
			"flag_reasons": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Available flag reasons for the channel.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Computed:    true,
							Description: "Flag reason code (e.g., SAFETY, DISPUTE).",
						},
						"label": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable label.",
						},
						"severity": schema.StringAttribute{
							Computed:    true,
							Description: "Severity level: LOW, MEDIUM, HIGH, or URGENT.",
						},
					},
				},
			},
			"auto_detection": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Auto-detection patterns configured for the channel.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"patterns": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Text patterns to detect in messages.",
						},
						"flag_reason": schema.StringAttribute{
							Computed:    true,
							Description: "Flag reason code applied when patterns match.",
						},
						"auto_flag": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether matching messages are automatically flagged.",
						},
					},
				},
			},
			"rules": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Escalation rules configured for the channel.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Rule identifier.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Rule name.",
						},
						"trigger": schema.StringAttribute{
							Computed:    true,
							Description: "Trigger event (e.g., PARTICIPANT_FLAG).",
						},
						"conditions": schema.StringAttribute{
							Computed:    true,
							Description: "JSON Logic condition expression.",
						},
						"actions": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Actions executed when the rule fires.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Computed:    true,
										Description: "Action type (CREATE_ISSUE, NOTIFY_PARTICIPANTS, WEBHOOK).",
									},
									"config": schema.StringAttribute{
										Computed:    true,
										Description: "JSON-encoded action configuration.",
									},
								},
							},
						},
						"priority": schema.Int64Attribute{
							Computed:    true,
							Description: "Rule evaluation priority.",
						},
						"is_active": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the rule is active.",
						},
						"cooldown_minutes": schema.Int64Attribute{
							Computed:    true,
							Description: "Minutes between consecutive rule firings.",
						},
					},
				},
			},
		},
	}
}

func (d *ThreadsEscalationConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*platformxe.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected *platformxe.Client")
		return
	}
	d.client = client
}

func (d *ThreadsEscalationConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ThreadsEscalationConfigDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.Threads.GetEscalationConfig(config.ChannelID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Escalation Config Failed", err.Error())
		return
	}

	// Parse the rules map into structured state.
	// The EscalationConfig type has Rules as map[string]interface{} — parse sub-keys.
	if result.Rules != nil {
		if flagReasons, ok := result.Rules["flagReasons"].([]interface{}); ok {
			config.FlagReasons = make([]EscalationFlagReasonItem, 0, len(flagReasons))
			for _, fr := range flagReasons {
				frm, ok := fr.(map[string]interface{})
				if !ok {
					continue
				}
				item := EscalationFlagReasonItem{}
				if v, ok := frm["code"].(string); ok {
					item.Code = types.StringValue(v)
				}
				if v, ok := frm["label"].(string); ok {
					item.Label = types.StringValue(v)
				}
				if v, ok := frm["severity"].(string); ok {
					item.Severity = types.StringValue(v)
				}
				config.FlagReasons = append(config.FlagReasons, item)
			}
		}

		if autoDetection, ok := result.Rules["autoDetection"].([]interface{}); ok {
			config.AutoDetection = make([]EscalationAutoDetectionItem, 0, len(autoDetection))
			for _, ad := range autoDetection {
				adm, ok := ad.(map[string]interface{})
				if !ok {
					continue
				}
				item := EscalationAutoDetectionItem{}
				if patterns, ok := adm["patterns"].([]interface{}); ok {
					strs := make([]string, 0, len(patterns))
					for _, p := range patterns {
						if s, ok := p.(string); ok {
							strs = append(strs, s)
						}
					}
					list, _ := types.ListValueFrom(ctx, types.StringType, strs)
					item.Patterns = list
				}
				if v, ok := adm["flagReason"].(string); ok {
					item.FlagReason = types.StringValue(v)
				}
				if v, ok := adm["autoFlag"].(bool); ok {
					item.AutoFlag = types.BoolValue(v)
				}
				config.AutoDetection = append(config.AutoDetection, item)
			}
		}

		if rules, ok := result.Rules["rules"].([]interface{}); ok {
			config.Rules = make([]EscalationRuleItem, 0, len(rules))
			for _, r := range rules {
				rm, ok := r.(map[string]interface{})
				if !ok {
					continue
				}
				rule := EscalationRuleItem{}
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
				if cond, ok := rm["conditions"]; ok {
					b, _ := json.Marshal(cond)
					rule.Conditions = types.StringValue(string(b))
				}
				if actions, ok := rm["actions"].([]interface{}); ok {
					rule.Actions = make([]EscalationRuleActionItem, 0, len(actions))
					for _, a := range actions {
						am, ok := a.(map[string]interface{})
						if !ok {
							continue
						}
						action := EscalationRuleActionItem{}
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
				config.Rules = append(config.Rules, rule)
			}
		}
	}

	// Ensure non-nil slices for Terraform state.
	if config.FlagReasons == nil {
		config.FlagReasons = []EscalationFlagReasonItem{}
	}
	if config.AutoDetection == nil {
		config.AutoDetection = []EscalationAutoDetectionItem{}
	}
	if config.Rules == nil {
		config.Rules = []EscalationRuleItem{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
