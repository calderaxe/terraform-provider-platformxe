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
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &TemplateResource{}
var _ resource.ResourceWithConfigure = &TemplateResource{}
var _ resource.ResourceWithImportState = &TemplateResource{}

// ImportState imports a template by its ID; Read then fills body_blocks_json
// from the API.
func (r *TemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type TemplateResource struct {
	client *platformxe.Client
}

type TemplateResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Subject        types.String `tfsdk:"subject"`
	BodyBlocksJSON types.String `tfsdk:"body_blocks_json"`
}

// decodeBodyBlocks parses body_blocks_json into the API's bodyBlocks array.
func decodeBodyBlocks(raw string) ([]map[string]interface{}, error) {
	var blocks []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &blocks); err != nil {
		return nil, fmt.Errorf("body_blocks_json must be a JSON array of blocks: %w", err)
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("body_blocks_json must contain at least one block")
	}
	return blocks, nil
}

// sameBodyBlocks compares two block lists as JSON values, so key order,
// whitespace and number formatting never show up as a diff.
func sameBodyBlocks(a, b []map[string]interface{}) bool {
	ja, errA := json.Marshal(a)
	jb, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	var va, vb interface{}
	if json.Unmarshal(ja, &va) != nil || json.Unmarshal(jb, &vb) != nil {
		return false
	}
	return reflect.DeepEqual(va, vb)
}

func NewTemplatesTemplateResource() resource.Resource {
	return &TemplateResource{}
}

func (r *TemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_templates_template"
}

func (r *TemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe email/message template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Template identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Template name.",
			},
			"subject": schema.StringAttribute{
				Required:    true,
				Description: "Email subject line (supports template variables).",
			},
			"body_blocks_json": schema.StringAttribute{
				Required: true,
				Description: "Template body as a JSON array of blocks (use jsonencode). Each block has a " +
					"type — text, heading, button, image, divider or spacer — and the fields that type uses: " +
					"content, level, label, url, src, alt, height.",
			},
		},
	}
}

func (r *TemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	blocks, err := decodeBodyBlocks(plan.BodyBlocksJSON.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid body_blocks_json", err.Error())
		return
	}

	input := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"subject":    plan.Subject.ValueString(),
		"bodyBlocks": blocks,
	}

	result, err := r.client.Templates.Create(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Template Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Templates.Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Template Failed", err.Error())
		return
	}

	state.Name = types.StringValue(result.Name)
	state.Subject = types.StringValue(result.Subject)

	// Keep the configured JSON text unless the blocks really changed
	// remotely; then record the API's blocks so the plan shows the drift.
	current, decodeErr := decodeBodyBlocks(state.BodyBlocksJSON.ValueString())
	if decodeErr != nil || !sameBodyBlocks(current, result.BodyBlocks) {
		remote, err := json.Marshal(result.BodyBlocks)
		if err != nil {
			resp.Diagnostics.AddError("Read Template Failed", err.Error())
			return
		}
		state.BodyBlocksJSON = types.StringValue(string(remote))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *TemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	blocks, err := decodeBodyBlocks(plan.BodyBlocksJSON.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid body_blocks_json", err.Error())
		return
	}

	input := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"subject":    plan.Subject.ValueString(),
		"bodyBlocks": blocks,
	}

	_, err = r.client.Templates.Update(plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Update Template Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Templates.Delete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Template Failed", err.Error())
		return
	}
}
