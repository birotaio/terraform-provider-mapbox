package datasources

import (
	"context"
	"fmt"

	"github.com/birotaio/terraform-provider-mapbox/internal/mapbox"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &StyleDataSource{}

func NewStyleDataSource() datasource.DataSource {
	return &StyleDataSource{}
}

type StyleDataSource struct {
	client *mapbox.Client
}

type StyleDataSourceModel struct {
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

func (d *StyleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_style"
}

func (d *StyleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read information about an existing Mapbox style.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the style to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the style.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The Mapbox style specification version.",
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "A JSON string containing arbitrary style metadata.",
				Computed:            true,
			},
			"sources": schema.StringAttribute{
				MarkdownDescription: "A JSON string defining the map data sources.",
				Computed:            true,
			},
			"layers": schema.StringAttribute{
				MarkdownDescription: "A JSON string defining the map layers.",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "The visibility of the style: `public` or `private`.",
				Computed:            true,
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "The username of the style owner.",
				Computed:            true,
			},
			"sprite": schema.StringAttribute{
				MarkdownDescription: "The sprite URL.",
				Computed:            true,
			},
			"glyphs": schema.StringAttribute{
				MarkdownDescription: "The font glyphs URL.",
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

func (d *StyleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mapbox.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mapbox.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *StyleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StyleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	style, err := d.client.GetStyle(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Mapbox Style", fmt.Sprintf("Unable to read style %s: %s", data.ID.ValueString(), err))
		return
	}

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
	} else {
		data.Visibility = types.StringNull()
	}

	if len(style.Metadata) > 0 && string(style.Metadata) != "null" {
		data.Metadata = types.StringValue(string(style.Metadata))
	} else {
		data.Metadata = types.StringNull()
	}

	if len(style.Sources) > 0 && string(style.Sources) != "null" {
		data.Sources = types.StringValue(string(style.Sources))
	} else {
		data.Sources = types.StringNull()
	}

	if len(style.Layers) > 0 && string(style.Layers) != "null" {
		data.Layers = types.StringValue(string(style.Layers))
	} else {
		data.Layers = types.StringNull()
	}

	tflog.Trace(ctx, "read mapbox style data source", map[string]any{"id": style.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
