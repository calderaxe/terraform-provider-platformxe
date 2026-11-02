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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &StorageProcessorResource{}
var _ resource.ResourceWithConfigure = &StorageProcessorResource{}

type StorageProcessorResource struct {
	client *platformxe.Client
}

type StorageProcessorResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	Provider             types.String `tfsdk:"provider"`
	ModerationEnabled    types.Bool   `tfsdk:"moderation_enabled"`
	ModerationAutoReject types.Bool   `tfsdk:"moderation_auto_reject"`
	BlockedMimeTypes     types.List   `tfsdk:"blocked_mime_types"`
	MaxImageSizeMb       types.Int64  `tfsdk:"max_image_size_mb"`
	MaxDocumentSizeMb    types.Int64  `tfsdk:"max_document_size_mb"`
	MaxVideoSizeMb       types.Int64  `tfsdk:"max_video_size_mb"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
}

func NewStorageProcessorResource() resource.Resource {
	return &StorageProcessorResource{}
}

func (r *StorageProcessorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_processor"
}

func (r *StorageProcessorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the PlatformXe storage processor configuration for media and document file operations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Processor identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the storage processor is enabled. Defaults to true.",
			},
			"provider": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("cloudinary"),
				Description: "Storage provider. Defaults to \"cloudinary\".",
			},
			"moderation_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether content moderation is enabled. Defaults to true.",
			},
			"moderation_auto_reject": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to automatically reject moderated content. Defaults to false.",
			},
			"blocked_mime_types": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of MIME types to block from upload.",
			},
			"max_image_size_mb": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Description: "Maximum image file size in MB. Defaults to 10.",
			},
			"max_document_size_mb": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(25),
				Description: "Maximum document file size in MB. Defaults to 25.",
			},
			"max_video_size_mb": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Description: "Maximum video file size in MB. Defaults to 100.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the processor was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the processor was last updated.",
			},
		},
	}
}

func (r *StorageProcessorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StorageProcessorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StorageProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":              plan.Enabled.ValueBool(),
		"provider":             plan.Provider.ValueString(),
		"moderationEnabled":    plan.ModerationEnabled.ValueBool(),
		"moderationAutoReject": plan.ModerationAutoReject.ValueBool(),
		"maxImageSizeMb":       plan.MaxImageSizeMb.ValueInt64(),
		"maxDocumentSizeMb":    plan.MaxDocumentSizeMb.ValueInt64(),
		"maxVideoSizeMb":       plan.MaxVideoSizeMb.ValueInt64(),
	}

	if !plan.BlockedMimeTypes.IsNull() && !plan.BlockedMimeTypes.IsUnknown() {
		var mimeTypes []string
		resp.Diagnostics.Append(plan.BlockedMimeTypes.ElementsAs(ctx, &mimeTypes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["blockedMimeTypes"] = mimeTypes
	}

	result, err := r.client.Storage.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Storage Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *StorageProcessorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StorageProcessorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Storage.GetProcessor()
	if err != nil {
		resp.Diagnostics.AddError("Read Storage Processor Failed", err.Error())
		return
	}

	state.Enabled = types.BoolValue(result.Enabled)
	mapProcessorConfigToComputedState(result, &state.ID, &state.CreatedAt, &state.UpdatedAt)

	cfg := result.Config
	if v, ok := cfg["provider"].(string); ok {
		state.Provider = types.StringValue(v)
	}
	if v, ok := cfg["moderationEnabled"].(bool); ok {
		state.ModerationEnabled = types.BoolValue(v)
	}
	if v, ok := cfg["moderationAutoReject"].(bool); ok {
		state.ModerationAutoReject = types.BoolValue(v)
	}
	if v, ok := cfg["blockedMimeTypes"].([]interface{}); ok {
		list, diags := types.ListValueFrom(ctx, types.StringType, interfaceSliceToStringSlice(v))
		resp.Diagnostics.Append(diags...)
		state.BlockedMimeTypes = list
	}
	if v, ok := cfg["maxImageSizeMb"].(float64); ok {
		state.MaxImageSizeMb = types.Int64Value(int64(v))
	}
	if v, ok := cfg["maxDocumentSizeMb"].(float64); ok {
		state.MaxDocumentSizeMb = types.Int64Value(int64(v))
	}
	if v, ok := cfg["maxVideoSizeMb"].(float64); ok {
		state.MaxVideoSizeMb = types.Int64Value(int64(v))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *StorageProcessorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StorageProcessorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"enabled":              plan.Enabled.ValueBool(),
		"provider":             plan.Provider.ValueString(),
		"moderationEnabled":    plan.ModerationEnabled.ValueBool(),
		"moderationAutoReject": plan.ModerationAutoReject.ValueBool(),
		"maxImageSizeMb":       plan.MaxImageSizeMb.ValueInt64(),
		"maxDocumentSizeMb":    plan.MaxDocumentSizeMb.ValueInt64(),
		"maxVideoSizeMb":       plan.MaxVideoSizeMb.ValueInt64(),
	}

	if !plan.BlockedMimeTypes.IsNull() && !plan.BlockedMimeTypes.IsUnknown() {
		var mimeTypes []string
		resp.Diagnostics.Append(plan.BlockedMimeTypes.ElementsAs(ctx, &mimeTypes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["blockedMimeTypes"] = mimeTypes
	}

	result, err := r.client.Storage.UpdateProcessor(input)
	if err != nil {
		resp.Diagnostics.AddError("Update Storage Processor Failed", err.Error())
		return
	}

	mapProcessorConfigToComputedState(result, &plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *StorageProcessorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.Storage.UpdateProcessor(map[string]interface{}{"enabled": false})
	if err != nil {
		resp.Diagnostics.AddError("Disable Storage Processor Failed", err.Error())
		return
	}
}
