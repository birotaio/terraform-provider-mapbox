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

var _ datasource.DataSource = &TokenDataSource{}

func NewTokenDataSource() datasource.DataSource {
	return &TokenDataSource{}
}

type TokenDataSource struct {
	client *mapbox.Client
}

type TokenDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Note        types.String `tfsdk:"note"`
	Scopes      types.Set    `tfsdk:"scopes"`
	AllowedUrls types.List   `tfsdk:"allowed_urls"`
	Token       types.String `tfsdk:"token"`
	Usage       types.String `tfsdk:"usage"`
	Default     types.Bool   `tfsdk:"default"`
	Created     types.String `tfsdk:"created"`
	Modified    types.String `tfsdk:"modified"`
}

func (d *TokenDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

func (d *TokenDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read information about an existing Mapbox access token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the token to look up.",
				Required:            true,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "A human-readable description of the token.",
				Computed:            true,
			},
			"scopes": schema.SetAttribute{
				MarkdownDescription: "The set of scopes granted to the token.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"allowed_urls": schema.ListAttribute{
				MarkdownDescription: "A list of URLs that the token is restricted to.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The actual token string. This value is sensitive.",
				Computed:            true,
				Sensitive:           true,
			},
			"usage": schema.StringAttribute{
				MarkdownDescription: "The token type: `pk` (public) or `sk` (secret).",
				Computed:            true,
			},
			"default": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the default token for the account.",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The ISO 8601 timestamp when the token was created.",
				Computed:            true,
			},
			"modified": schema.StringAttribute{
				MarkdownDescription: "The ISO 8601 timestamp when the token was last modified.",
				Computed:            true,
			},
		},
	}
}

func (d *TokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TokenDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token, err := d.client.GetToken(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Mapbox Token", fmt.Sprintf("Unable to read token %s: %s", data.ID.ValueString(), err))
		return
	}

	data.ID = types.StringValue(token.ID)
	data.Note = types.StringValue(token.Note)
	data.Usage = types.StringValue(token.Usage)
	data.Default = types.BoolValue(token.Default)
	data.Created = types.StringValue(token.Created)
	data.Modified = types.StringValue(token.Modified)

	if token.TokenString != "" {
		data.Token = types.StringValue(token.TokenString)
	} else {
		data.Token = types.StringNull()
	}

	scopeValues, diags := types.SetValueFrom(ctx, types.StringType, token.Scopes)
	resp.Diagnostics.Append(diags...)
	data.Scopes = scopeValues

	if len(token.AllowedUrls) > 0 {
		urlValues, diags := types.ListValueFrom(ctx, types.StringType, token.AllowedUrls)
		resp.Diagnostics.Append(diags...)
		data.AllowedUrls = urlValues
	} else {
		data.AllowedUrls = types.ListNull(types.StringType)
	}

	tflog.Trace(ctx, "read mapbox token data source", map[string]any{"id": token.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
