package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/birotaio/terraform-provider-mapbox/internal/mapbox"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Version    types.Int64  `tfsdk:"version"`
	Metadata   types.String `tfsdk:"metadata"`
	Sources    types.String `tfsdk:"sources"`
	Layers     types.String `tfsdk:"layers"`
	Visibility types.String `tfsdk:"visibility"`
	Owner      types.String `tfsdk:"owner"`
	Sprite     types.String `tfsdk:"sprite"`
	Glyphs     types.String `tfsdk:"glyphs"`
	Created    types.String `tfsdk:"created"`
	Modified   types.String `tfsdk:"modified"`
	Draft      types.Bool   `tfsdk:"draft"`
	Protected  types.Bool   `tfsdk:"protected"`
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
			"metadata": schema.StringAttribute{
				MarkdownDescription: "A JSON string containing arbitrary style metadata.",
				Optional:            true,
			},
			"sources": schema.StringAttribute{
				MarkdownDescription: "A JSON string defining the map data sources. Maximum 15 sources. " +
					"Use `jsonencode()` in your Terraform configuration to construct this value.",
				Optional: true,
			},
			"layers": schema.StringAttribute{
				MarkdownDescription: "A JSON string defining the map layers and rendering rules. " +
					"Use `jsonencode()` in your Terraform configuration to construct this value.",
				Optional: true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "The visibility of the style: `public` or `private`.",
				Optional:            true,
				Computed:            true,
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "The username of the style owner.",
				Computed:            true,
			},
			"sprite": schema.StringAttribute{
				MarkdownDescription: "The sprite URL, automatically set by the Mapbox API.",
				Computed:            true,
			},
			"glyphs": schema.StringAttribute{
				MarkdownDescription: "The font glyphs URL, automatically set by the Mapbox API.",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The ISO 8601 timestamp when the style was created.",
				Computed:            true,
			},
			"modified": schema.StringAttribute{
				MarkdownDescription: "The ISO 8601 timestamp when the style was last modified.",
				Computed:            true,
			},
			"draft": schema.BoolAttribute{
				MarkdownDescription: "Whether the style has unpublished changes.",
				Computed:            true,
			},
			"protected": schema.BoolAttribute{
				MarkdownDescription: "Whether the style is protected from editing.",
				Computed:            true,
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
	var data StyleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := mapbox.CreateStyleRequest{
		Version: int(data.Version.ValueInt64()),
		Name:    data.Name.ValueString(),
	}

	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		createReq.Visibility = data.Visibility.ValueString()
	}

	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		createReq.Metadata = json.RawMessage(data.Metadata.ValueString())
	}

	if !data.Sources.IsNull() && !data.Sources.IsUnknown() {
		createReq.Sources = json.RawMessage(data.Sources.ValueString())
	} else {
		createReq.Sources = json.RawMessage(`{}`)
	}

	if !data.Layers.IsNull() && !data.Layers.IsUnknown() {
		createReq.Layers = json.RawMessage(data.Layers.ValueString())
	} else {
		createReq.Layers = json.RawMessage(`[]`)
	}

	style, err := r.client.CreateStyle(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Mapbox Style", fmt.Sprintf("Unable to create style: %s", err))
		return
	}

	diags := mapStyleToState(style, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created mapbox style", map[string]any{"id": style.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StyleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StyleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	style, err := r.client.GetStyle(ctx, data.ID.ValueString())
	if err != nil {
		if mapbox.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Mapbox Style", fmt.Sprintf("Unable to read style %s: %s", data.ID.ValueString(), err))
		return
	}

	diags := mapStyleToState(style, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StyleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StyleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := mapbox.UpdateStyleRequest{
		Version: int(data.Version.ValueInt64()),
		Name:    data.Name.ValueString(),
		Owner:   r.client.Username,
	}

	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		updateReq.Visibility = data.Visibility.ValueString()
	}

	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		updateReq.Metadata = json.RawMessage(data.Metadata.ValueString())
	}

	if !data.Sources.IsNull() && !data.Sources.IsUnknown() {
		updateReq.Sources = json.RawMessage(data.Sources.ValueString())
	} else {
		updateReq.Sources = json.RawMessage(`{}`)
	}

	if !data.Layers.IsNull() && !data.Layers.IsUnknown() {
		updateReq.Layers = json.RawMessage(data.Layers.ValueString())
	} else {
		updateReq.Layers = json.RawMessage(`[]`)
	}

	if !data.Sprite.IsNull() && !data.Sprite.IsUnknown() {
		updateReq.Sprite = data.Sprite.ValueString()
	}

	if !data.Glyphs.IsNull() && !data.Glyphs.IsUnknown() {
		updateReq.Glyphs = data.Glyphs.ValueString()
	}

	style, err := r.client.UpdateStyle(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Mapbox Style", fmt.Sprintf("Unable to update style %s: %s", data.ID.ValueString(), err))
		return
	}

	diags := mapStyleToState(style, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated mapbox style", map[string]any{"id": style.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
func mapStyleToState(style *mapbox.Style, data *StyleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(style.ID)
	data.Name = types.StringValue(style.Name)
	data.Version = types.Int64Value(int64(style.Version))
	data.Owner = types.StringValue(style.Owner)
	data.Sprite = types.StringValue(style.Sprite)
	data.Glyphs = types.StringValue(style.Glyphs)
	data.Created = types.StringValue(style.Created)
	data.Modified = types.StringValue(style.Modified)
	data.Draft = types.BoolValue(style.Draft)
	data.Protected = types.BoolValue(style.Protected)

	if style.Visibility != "" {
		data.Visibility = types.StringValue(style.Visibility)
	}

	if len(style.Metadata) > 0 && string(style.Metadata) != "null" {
		data.Metadata = types.StringValue(string(style.Metadata))
	}

	if len(style.Sources) > 0 && string(style.Sources) != "null" {
		data.Sources = types.StringValue(string(style.Sources))
	}

	if len(style.Layers) > 0 && string(style.Layers) != "null" {
		data.Layers = types.StringValue(string(style.Layers))
	}

	return diags
}
