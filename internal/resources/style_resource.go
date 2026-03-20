package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/birotaio/terraform-provider-mapbox/internal/mapbox"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &StyleResource{}
	_ resource.ResourceWithImportState = &StyleResource{}
)

func NewStyleResource() resource.Resource {
	return &StyleResource{}
}

type StyleResource struct {
	client *mapbox.Client
}

type StyleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Version         types.Int64  `tfsdk:"version"`
	InitialMetadata types.String `tfsdk:"initial_metadata"`
	Metadata        types.String `tfsdk:"metadata"`
	InitialSources  types.String `tfsdk:"initial_sources"`
	Sources         types.String `tfsdk:"sources"`
	InitialLayers   types.String `tfsdk:"initial_layers"`
	Layers          types.String `tfsdk:"layers"`
	Visibility      types.String `tfsdk:"visibility"`
	Owner           types.String `tfsdk:"owner"`
	Sprite          types.String `tfsdk:"sprite"`
	Glyphs          types.String `tfsdk:"glyphs"`
	Draft           types.Bool   `tfsdk:"draft"`
	Protected       types.Bool   `tfsdk:"protected"`
	FolderId        types.String `tfsdk:"folder_id"`
}

func (r *StyleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_style"
}

func (r *StyleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Mapbox style. Styles define the visual appearance of maps, " +
			"including sources, layers, and rendering rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the style.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the style.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The Mapbox style specification version. Defaults to `8`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(8),
			},
			"initial_metadata": schema.StringAttribute{
				MarkdownDescription: "A JSON string used as the initial metadata when creating the style. " +
					"Cannot be used together with `metadata`. Changing this value after creation has no effect.",
				Optional:  true,
				WriteOnly: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("metadata")),
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "A JSON string containing arbitrary style metadata. " +
					"Cannot be used together with `initial_metadata`.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("initial_metadata")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"initial_sources": schema.StringAttribute{
				MarkdownDescription: "A JSON string used as the initial sources when creating the style. " +
					"Changing this value after creation has no effect.",
				Optional:  true,
				WriteOnly: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("sources")),
				},
			},
			"sources": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "A JSON string defining the map data sources. Maximum 15 sources. " +
					"Use `jsonencode()` in your Terraform configuration to construct this value.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("initial_sources")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"initial_layers": schema.StringAttribute{
				MarkdownDescription: "A JSON string used as the initial layers when creating the style. " +
					"Cannot be used together with `layers`. Changing this value after creation has no effect.",
				Optional:  true,
				WriteOnly: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("layers")),
				},
			},
			"layers": schema.StringAttribute{
				MarkdownDescription: "A JSON string defining the map layers and rendering rules. " +
					"Use `jsonencode()` in your Terraform configuration to construct this value. " +
					"Cannot be used together with `initial_layers`.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("initial_layers")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "The visibility of the style: `public` or `private`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("private"),
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "The username of the style owner.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sprite": schema.StringAttribute{
				MarkdownDescription: "The sprite URL, automatically set by the Mapbox API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"glyphs": schema.StringAttribute{
				MarkdownDescription: "The font glyphs URL, automatically set by the Mapbox API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"draft": schema.BoolAttribute{
				MarkdownDescription: "Whether the style has unpublished changes.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"protected": schema.BoolAttribute{
				MarkdownDescription: "Whether the style is protected from editing.",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the folder containing the style. Leave unset not to control the folder. Use `styleroot` ID to put a the root folder.",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("styleroot"),
			},
		},
	}
}

func (r *StyleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mapbox.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mapbox.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *StyleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config, data StyleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := mapbox.CreateStyleRequest{
		Version:    int(data.Version.ValueInt64()),
		Name:       data.Name.ValueString(),
		Visibility: data.Visibility.ValueString(),
	}

	if !config.InitialSources.IsNull() && !config.InitialSources.IsUnknown() {
		createReq.Sources = json.RawMessage(config.InitialSources.ValueString())
	} else {
		createReq.Sources = json.RawMessage(data.Sources.ValueString())
	}

	// Bootstrap pattern: initial_metadata seeds the value on creation; metadata manages it ongoing.
	if !config.InitialMetadata.IsNull() && !config.InitialMetadata.IsUnknown() {
		createReq.Metadata = json.RawMessage(config.InitialMetadata.ValueString())
	} else {
		createReq.Metadata = json.RawMessage(data.Metadata.ValueString())
	}

	// Bootstrap pattern: initial_layers seeds the value on creation; layers manages it ongoing.
	if !config.InitialLayers.IsNull() && !config.InitialLayers.IsUnknown() {
		createReq.Layers = json.RawMessage(config.InitialLayers.ValueString())
	} else {
		createReq.Layers = json.RawMessage(data.Layers.ValueString())
	}

	style, err := r.client.CreateStyle(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Mapbox Style", fmt.Sprintf("Unable to create style: %s", err))
		return
	}

	if config.Protected.ValueBool() {
		if err := r.client.SetStyleProtected(ctx, style.ID, true); err != nil {
			resp.Diagnostics.AddError("Error Setting Style Protected", fmt.Sprintf("Style was created but failed to set protected: %s", err))
		} else {
			style.Protected = true
		}
	}

	diags := mapStyleToState(style, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.FolderId.IsUnknown() && !config.FolderId.IsNull() {
		folderId := config.FolderId.ValueString()
		if err := r.client.SetStyleFolderId(ctx, style.ID, folderId); err != nil {
			resp.Diagnostics.AddError("Error Setting Style Folder", fmt.Sprintf("Style was created but failed to set folder ID to %s: %s", folderId, err))
		} else {
			data.FolderId = types.StringValue(folderId)
		}
	}

	tflog.Trace(ctx, "created mapbox style", map[string]any{"id": style.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StyleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var reqState, respState StyleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &reqState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	style, err := r.client.GetStyle(ctx, reqState.ID.ValueString())
	if err != nil {
		if mapbox.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Mapbox Style", fmt.Sprintf("Unable to read style %s: %s", reqState.ID.ValueString(), err))
		return
	}

	diags := mapStyleToState(style, &respState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folderId, err := r.client.GetStyleFolderId(ctx, style.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Getting Style Folder", fmt.Sprintf("Style was read but failed to get folder ID: %s", err))
	} else {
		respState.FolderId = types.StringValue(folderId)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &respState)...)
}

func (r *StyleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var currentState, plan, config StyleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use Diff to detect which attributes actually changed between state and plan.
	diffs, err := req.Plan.Raw.Diff(req.State.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Error Computing Diff", fmt.Sprintf("Unable to diff plan and state: %s", err))
		return
	}

	styleFields := map[string]bool{
		"name": true, "version": true, "visibility": true,
		"sources": true, "metadata": true, "layers": true,
	}
	styleChanged := false
	for _, d := range diffs {
		steps := d.Path.Steps()
		if len(steps) > 0 {
			if attrName, ok := steps[0].(tftypes.AttributeName); ok && styleFields[string(attrName)] {
				styleChanged = true
				break
			}
		}
	}

	var style *mapbox.Style
	if styleChanged {
		updateReq := mapbox.UpdateStyleRequest{
			Version:    int(plan.Version.ValueInt64()),
			Name:       plan.Name.ValueString(),
			Owner:      r.client.Username,
			Visibility: plan.Visibility.ValueString(),
		}

		// Only update sources when the user has explicitly defined it in their configuration.
		if !config.Sources.IsNull() && !config.Sources.IsUnknown() {
			updateReq.Sources = json.RawMessage(config.Sources.ValueString())
		} else {
			updateReq.Sources = json.RawMessage(plan.Sources.ValueString())
		}

		// Only update metadata when the user has explicitly defined it in their configuration.
		if !config.Metadata.IsNull() && !config.Metadata.IsUnknown() {
			updateReq.Metadata = json.RawMessage(config.Metadata.ValueString())
		} else {
			updateReq.Metadata = json.RawMessage(plan.Metadata.ValueString())
		}
		// Only update layers when the user has explicitly defined it in their configuration.
		if !config.Layers.IsNull() && !config.Layers.IsUnknown() {
			updateReq.Layers = json.RawMessage(config.Layers.ValueString())
		} else {
			updateReq.Layers = json.RawMessage(plan.Layers.ValueString())
		}

		var err error
		style, err = r.client.UpdateStyle(ctx, plan.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Mapbox Style", fmt.Sprintf("Unable to update style %s: %s", plan.ID.ValueString(), err))
			return
		}
	} else {
		// No style fields changed; read current state from the API.
		var err error
		style, err = r.client.GetStyle(ctx, plan.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Mapbox Style", fmt.Sprintf("Unable to read style %s: %s", plan.ID.ValueString(), err))
			return
		}
	}

	if style.Protected != config.Protected.ValueBool() {
		if err := r.client.SetStyleProtected(ctx, style.ID, config.Protected.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Error Updating Style Protected Setting", fmt.Sprintf("Style was updated but failed to update protected setting: %s", err))
		} else {
			style.Protected = config.Protected.ValueBool()
		}
	}

	diags := mapStyleToState(style, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if currentState.FolderId.ValueString() != config.FolderId.ValueString() {
		folderId := config.FolderId.ValueString()
		if err := r.client.SetStyleFolderId(ctx, style.ID, folderId); err != nil {
			resp.Diagnostics.AddError("Error Updating Style Folder", fmt.Sprintf("Style was updated but failed to update folder ID to %s: %s", folderId, err))
		} else {
			plan.FolderId = types.StringValue(folderId)
		}
	}

	tflog.Trace(ctx, "updated mapbox style", map[string]any{"id": style.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StyleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StyleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStyle(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Mapbox Style", fmt.Sprintf("Unable to delete style %s: %s", data.ID.ValueString(), err))
		return
	}

	tflog.Trace(ctx, "deleted mapbox style", map[string]any{"id": data.ID.ValueString()})
}

func (r *StyleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapStyleToState maps a Mapbox API Style response to the Terraform resource model.
func mapStyleToState(style *mapbox.Style, state *StyleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(style.ID)
	state.Name = types.StringValue(style.Name)
	state.Version = types.Int64Value(int64(style.Version))
	state.Owner = types.StringValue(style.Owner)
	state.Sprite = types.StringValue(style.Sprite)
	state.Sources = types.StringValue(string(style.Sources))
	state.Glyphs = types.StringValue(style.Glyphs)
	state.Draft = types.BoolValue(style.Draft)
	state.Protected = types.BoolValue(style.Protected)
	state.Metadata = types.StringValue(string(style.Metadata))
	state.Layers = types.StringValue(string(style.Layers))
	state.Visibility = types.StringValue(style.Visibility)

	return diags
}
