// =============================================================================
// Copyright 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	platformxe "github.com/calderaxe/platformxe-go"
)

var _ resource.Resource = &PermissionsModuleResource{}
var _ resource.ResourceWithConfigure = &PermissionsModuleResource{}

type PermissionsModuleResource struct {
	client *platformxe.Client
}

type PermissionsModuleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	App         types.String `tfsdk:"app"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Paths       types.List   `tfsdk:"paths"`
}

func NewPermissionsModuleResource() resource.Resource {
	return &PermissionsModuleResource{}
}

func (r *PermissionsModuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_module"
}

func (r *PermissionsModuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PlatformXe permission module. Modules define groups of permission paths that can be assigned to roles. Note: the API does not support deleting modules — removing this resource from your configuration will only remove it from Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Module identifier (e.g., LT:BOOKINGS).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app": schema.StringAttribute{
				Required:    true,
				Description: "Owning application identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable module name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Module description.",
			},
			"paths": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Permission paths this module covers.",
			},
		},
	}
}

func (r *PermissionsModuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PermissionsModuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PermissionsModuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var paths []string
	resp.Diagnostics.Append(plan.Paths.ElementsAs(ctx, &paths, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]interface{}{
		"id":    plan.ID.ValueString(),
		"app":   plan.App.ValueString(),
		"name":  plan.Name.ValueString(),
		"paths": paths,
	}
	if !plan.Description.IsNull() {
		input["description"] = plan.Description.ValueString()
	}

	_, err := r.client.Permissions.RegisterModule(input)
	if err != nil {
		resp.Diagnostics.AddError("Create Permission Module Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PermissionsModuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PermissionsModuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Permissions.ListModules()
	if err != nil {
		resp.Diagnostics.AddError("Read Permission Module Failed", err.Error())
		return
	}

	var found *platformxe.PermissionModule
	for i := range result.Modules {
		if result.Modules[i].Key == state.ID.ValueString() {
			found = &result.Modules[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(found.Name)
	if found.Description != "" {
		state.Description = types.StringValue(found.Description)
	} else {
		state.Description = types.StringNull()
	}
	if found.Category != "" {
		state.App = types.StringValue(found.Category)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PermissionsModuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PermissionsModuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var paths []string
	resp.Diagnostics.Append(plan.Paths.ElementsAs(ctx, &paths, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-register the module to update it.
	input := map[string]interface{}{
		"id":    plan.ID.ValueString(),
		"app":   plan.App.ValueString(),
		"name":  plan.Name.ValueString(),
		"paths": paths,
	}
	if !plan.Description.IsNull() {
		input["description"] = plan.Description.ValueString()
	}

	_, err := r.client.Permissions.RegisterModule(input)
	if err != nil {
		resp.Diagnostics.AddError("Update Permission Module Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PermissionsModuleResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The PlatformXe API does not support deleting permission modules.
	// Removing from Terraform state only — the module remains registered on the platform.
	resp.Diagnostics.AddWarning(
		"Module Not Deleted From Platform",
		fmt.Sprintf("Permission modules cannot be deleted via the API. The resource has been removed from Terraform state but remains registered on PlatformXe."),
	)
}
