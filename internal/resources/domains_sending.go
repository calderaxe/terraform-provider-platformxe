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

var _ resource.Resource = &DomainResource{}
var _ resource.ResourceWithConfigure = &DomainResource{}

type DomainResource struct {
	client *platformxe.Client
}

// -- Terraform models --------------------------------------------------------

type DNSRecordModel struct {
	Type     types.String `tfsdk:"type"`
	Name     types.String `tfsdk:"name"`
	Value    types.String `tfsdk:"value"`
	Verified types.Bool   `tfsdk:"verified"`
}

type DomainResourceModel struct {
	ID         types.String     `tfsdk:"id"`
	Domain     types.String     `tfsdk:"domain"`
	Verified   types.Bool       `tfsdk:"verified"`
	DNSRecords []DNSRecordModel `tfsdk:"dns_records"`
}

func NewDomainsSendingResource() resource.Resource {
	return &DomainResource{}
}

func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains_sending"
}

func (r *DomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe sending domain for email delivery.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Domain identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain name to register for sending.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"verified": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the domain DNS records have been verified.",
			},
			"dns_records": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Required DNS records for domain verification.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "DNS record type (TXT, CNAME, MX).",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "DNS record name.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "DNS record value.",
						},
						"verified": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this individual record has been verified.",
						},
					},
				},
			},
		},
	}
}

func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// mapDomainToState populates the Terraform model from a typed SDK SendingDomain.
func mapDomainToState(d *platformxe.SendingDomain, state *DomainResourceModel) {
	state.ID = types.StringValue(d.ID)
	state.Domain = types.StringValue(d.Domain)
	state.Verified = types.BoolValue(d.VerifiedAt != nil)

	state.DNSRecords = make([]DNSRecordModel, 0, len(d.DNSRecords))
	for _, rec := range d.DNSRecords {
		state.DNSRecords = append(state.DNSRecords, DNSRecordModel{
			Type:     types.StringValue(rec.Type),
			Name:     types.StringValue(rec.Name),
			Value:    types.StringValue(rec.Value),
			Verified: types.BoolValue(rec.Verified),
		})
	}
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Domains.Add(plan.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Domain Failed", err.Error())
		return
	}

	mapDomainToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Domains.Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Domain Failed", err.Error())
		return
	}

	mapDomainToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Domains are immutable — the domain attribute has RequiresReplace.
	// This method should never be called, but is required by the interface.
	resp.Diagnostics.AddError("Update Not Supported", "Sending domains are immutable. Terraform will delete and recreate as needed.")
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Domains.Delete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Domain Failed", err.Error())
		return
	}
}
