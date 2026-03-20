package provider

import (
	"context"
	"os"
	"strings"

	"github.com/birotaio/terraform-provider-mapbox/internal/datasources"
	"github.com/birotaio/terraform-provider-mapbox/internal/mapbox"
	"github.com/birotaio/terraform-provider-mapbox/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &MapboxProvider{}

// MapboxProvider defines the provider implementation.
type MapboxProvider struct {
	version string
}

// MapboxProviderModel describes the provider data model.
type MapboxProviderModel struct {
	AccessToken types.String `tfsdk:"access_token"`
	Username    types.String `tfsdk:"username"`
	Fresh       types.Bool   `tfsdk:"fresh"`
}

func (p *MapboxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mapbox"
	resp.Version = p.version
}

func (p *MapboxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Mapbox provider allows you to manage Mapbox resources such as access tokens and styles.",
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Mapbox secret access token used for API authentication. " +
					"Must have the appropriate scopes for the resources being managed. " +
					"Can also be set with the `MAPBOX_ACCESS_TOKEN` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Mapbox account username. " +
					"Can also be set with the `MAPBOX_USERNAME` environment variable.",
				Optional: true,
			},
			"fresh": schema.BoolAttribute{
				MarkdownDescription: "Whether to use fresh data from the Mapbox API. " +
					"Can also be set with the `MAPBOX_FRESH` environment variable.",
				Optional: true,
			},
		},
	}
}

func (p *MapboxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MapboxProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessToken := data.AccessToken.ValueString()
	if accessToken == "" {
		accessToken = os.Getenv("MAPBOX_ACCESS_TOKEN")
	}
	if accessToken == "" {
		resp.Diagnostics.AddError(
			"Missing Mapbox Access Token",
			"The provider requires a Mapbox access token. Set it in the provider configuration "+
				"or via the MAPBOX_ACCESS_TOKEN environment variable.",
		)
	}

	username := data.Username.ValueString()
	if username == "" {
		username = os.Getenv("MAPBOX_USERNAME")
	}
	if username == "" {
		resp.Diagnostics.AddError(
			"Missing Mapbox Username",
			"The provider requires a Mapbox username. Set it in the provider configuration "+
				"or via the MAPBOX_USERNAME environment variable.",
		)
	}

	fresh := true
	if !data.Fresh.IsNull() && !data.Fresh.IsUnknown() {
		fresh = data.Fresh.ValueBool()
	}
	if envFresh := os.Getenv("MAPBOX_FRESH"); envFresh != "" {
		fresh = strings.ToLower(envFresh) == "true"
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := mapbox.NewClient(accessToken, username, fresh)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *MapboxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewTokenResource,
		resources.NewStyleResource,
	}
}

func (p *MapboxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewTokenDataSource,
		datasources.NewStyleDataSource,
	}
}

// New returns a factory function for the MapboxProvider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MapboxProvider{
			version: version,
		}
	}
}
