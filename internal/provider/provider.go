// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/livekit/protocol/auth"
)

var _ provider.Provider = &LivekitProvider{}
var _ provider.ProviderWithFunctions = &LivekitProvider{}

type LivekitProvider struct {
	version string
}

type LivekitProviderModel struct {
	ApiKey    types.String `tfsdk:"api_key"`
	ApiSecret types.String `tfsdk:"api_secret"`
}

func (p *LivekitProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "livekit"
	resp.Version = p.version
}

func (p *LivekitProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Livekit API Key. Can also be set via environment variable LIVEKIT_API_KEY",
				Optional:            true,
			},
			"api_secret": schema.StringAttribute{
				MarkdownDescription: "Livekit API Secret. Can also be set via environment variable LIVEKIT_API_SECRET",
				Optional:            true,
			},
		},
	}
}

func (p *LivekitProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data LivekitProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")

	if !data.ApiKey.IsNull() {
		apiKey = data.ApiKey.ValueString()
	}
	if !data.ApiSecret.IsNull() {
		apiSecret = data.ApiSecret.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError("Livekit api key missing",
			"The provider cannot create the Livekit client as there is a missing or empty value for the Livekit API key. "+
				"Set the api_key value in the configuration or use the LIVEKIT_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.")
	}

	if apiSecret == "" {
		resp.Diagnostics.AddError("Livekit api secret missing",
			"The provider cannot create the Livekit client as there is a missing or empty value for the Livekit API secret. "+
				"Set the api_secret value in the configuration or use the LIVEKIT_API_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	accessToken := auth.NewAccessToken(apiKey, apiSecret)

	resp.DataSourceData = accessToken
	resp.ResourceData = accessToken
}

func (p *LivekitProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *LivekitProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *LivekitProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LivekitProvider{
			version: version,
		}
	}
}
