package resources

import (
	"context"
	"fmt"

	"github.com/birotaio/terraform-provider-mapbox/internal/mapbox"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &TokenResource{}
	_ resource.ResourceWithImportState = &TokenResource{}
)

func NewTokenResource() resource.Resource {
	return &TokenResource{}
}

type TokenResource struct {
	client *mapbox.Client
}

type TokenResourceModel struct {
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

func (r *TokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

func (r *TokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Mapbox access token. Tokens provide API access to Mapbox services " +
			"with specific scopes and optional URL restrictions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the token.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "A human-readable description of the token.",
				Optional:            true,
			},
			"scopes": schema.SetAttribute{
				MarkdownDescription: "The set of scopes granted to the token (e.g., `styles:read`, `fonts:read`).",
				Required:            true,
				ElementType:         types.StringType,
			},
			"allowed_urls": schema.ListAttribute{
				MarkdownDescription: "A list of URLs that the token is restricted to. Maximum 100 URLs.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The actual token string. This value is sensitive and will not be displayed in logs.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *TokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := mapbox.CreateTokenRequest{
		Note: data.Note.ValueString(),
	}

	var scopes []string
	resp.Diagnostics.Append(data.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createReq.Scopes = scopes

	if !data.AllowedUrls.IsNull() && !data.AllowedUrls.IsUnknown() {
		var urls []string
		resp.Diagnostics.Append(data.AllowedUrls.ElementsAs(ctx, &urls, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.AllowedUrls = urls
	}

	token, err := r.client.CreateToken(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Mapbox Token", fmt.Sprintf("Unable to create token: %s", err))
		return
	}

	diags := mapTokenToState(ctx, token, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created mapbox token", map[string]any{"id": token.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.client.GetToken(ctx, data.ID.ValueString())
	if err != nil {
		if mapbox.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Mapbox Token", fmt.Sprintf("Unable to read token %s: %s", data.ID.ValueString(), err))
		return
	}

	diags := mapTokenToState(ctx, token, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := mapbox.UpdateTokenRequest{}

	if !data.Note.IsNull() && !data.Note.IsUnknown() {
		note := data.Note.ValueString()
		updateReq.Note = &note
	}

	var scopes []string
	resp.Diagnostics.Append(data.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq.Scopes = scopes

	if !data.AllowedUrls.IsNull() && !data.AllowedUrls.IsUnknown() {
		var urls []string
		resp.Diagnostics.Append(data.AllowedUrls.ElementsAs(ctx, &urls, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.AllowedUrls = urls
	}

	token, err := r.client.UpdateToken(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Mapbox Token", fmt.Sprintf("Unable to update token %s: %s", data.ID.ValueString(), err))
		return
	}

	diags := mapTokenToState(ctx, token, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated mapbox token", map[string]any{"id": token.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteToken(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Mapbox Token", fmt.Sprintf("Unable to delete token %s: %s", data.ID.ValueString(), err))
		return
	}

	tflog.Trace(ctx, "deleted mapbox token", map[string]any{"id": data.ID.ValueString()})
}

func (r *TokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapTokenToState maps a Mapbox API Token response to the Terraform resource model.
func mapTokenToState(ctx context.Context, token *mapbox.Token, data *TokenResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(token.ID)
	data.Note = types.StringValue(token.Note)
	data.Usage = types.StringValue(token.Usage)
	data.Default = types.BoolValue(token.Default)
	data.Created = types.StringValue(token.Created)
	data.Modified = types.StringValue(token.Modified)

	if token.TokenString != "" {
		data.Token = types.StringValue(token.TokenString)
	}

	scopeValues, d := types.SetValueFrom(ctx, types.StringType, token.Scopes)
	diags.Append(d...)
	data.Scopes = scopeValues

	if len(token.AllowedUrls) > 0 {
		urlValues, d := types.ListValueFrom(ctx, types.StringType, token.AllowedUrls)
		diags.Append(d...)
		data.AllowedUrls = urlValues
	} else if !data.AllowedUrls.IsNull() {
		data.AllowedUrls = types.ListNull(types.StringType)
	}

	return diags
}
