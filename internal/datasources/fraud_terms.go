// =============================================================================
// © 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ datasource.DataSource = &FraudTermsDataSource{}
var _ datasource.DataSourceWithConfigure = &FraudTermsDataSource{}

// FraudTermsDataSource reads the calling tenant's Fraud Detection Engine
// T&Cs acceptance state (Phase 6H). IaC plans can use this to assert
// acceptance at the current published version before depending on
// fraud:* / identity:verify resources, both of which are gated by
// `requireDetectionPackWithTerms`.
type FraudTermsDataSource struct {
	client *platformxe.Client
}

type FraudTermsDataSourceModel struct {
	Accepted        types.Bool   `tfsdk:"accepted"`
	AcceptedVersion types.String `tfsdk:"accepted_version"`
	AcceptedAt      types.String `tfsdk:"accepted_at"`
	AcceptedBy      types.String `tfsdk:"accepted_by"`
	CurrentVersion  types.String `tfsdk:"current_version"`
	StaleAcceptance types.Bool   `tfsdk:"stale_acceptance"`
}

func NewFraudTermsDataSource() datasource.DataSource {
	return &FraudTermsDataSource{}
}

func (d *FraudTermsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fraud_terms"
}

func (d *FraudTermsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the Fraud Detection Engine Terms & Conditions acceptance state for the calling organization. Useful for asserting click-through acceptance has been recorded before depending on fraud or identity verification resources.",
		Attributes: map[string]schema.Attribute{
			"accepted": schema.BoolAttribute{
				Computed:    true,
				Description: "True when the org has accepted at the current published version.",
			},
			"accepted_version": schema.StringAttribute{
				Computed:    true,
				Description: "Version that was accepted (empty when never accepted).",
			},
			"accepted_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO-8601 timestamp of acceptance (empty when never accepted).",
			},
			"accepted_by": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant user identifier that clicked through.",
			},
			"current_version": schema.StringAttribute{
				Computed:    true,
				Description: "The version currently in force.",
			},
			"stale_acceptance": schema.BoolAttribute{
				Computed:    true,
				Description: "True when an acceptance exists but the published version has been bumped — operator must re-accept.",
			},
		},
	}
}

func (d *FraudTermsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FraudTermsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Fraud.Terms.Status()
	if err != nil {
		resp.Diagnostics.AddError("Read Fraud Terms Status Failed", err.Error())
		return
	}

	var state FraudTermsDataSourceModel

	if accepted, ok := result["accepted"].(bool); ok {
		state.Accepted = types.BoolValue(accepted)
	}
	if v, ok := result["acceptedVersion"].(string); ok {
		state.AcceptedVersion = types.StringValue(v)
	} else {
		state.AcceptedVersion = types.StringValue("")
	}
	if v, ok := result["acceptedAt"].(string); ok {
		state.AcceptedAt = types.StringValue(v)
	} else {
		state.AcceptedAt = types.StringValue("")
	}
	if v, ok := result["acceptedBy"].(string); ok {
		state.AcceptedBy = types.StringValue(v)
	} else {
		state.AcceptedBy = types.StringValue("")
	}
	if v, ok := result["currentVersion"].(string); ok {
		state.CurrentVersion = types.StringValue(v)
	}
	if v, ok := result["staleAcceptance"].(bool); ok {
		state.StaleAcceptance = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
