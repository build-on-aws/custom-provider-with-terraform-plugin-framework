package buildonaws

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

var (
	_           provider.Provider = &buildOnAWSProvider{}
	commit, tag string            // populated by goreleaser
)

func New() provider.Provider {

	const (
		defaultVersion = "0.0.0"
		defaultCommit  = "devel"
	)

	return &buildOnAWSProvider{
		version: func() string {
			if len(tag) == 0 {
				return defaultVersion
			}
			return tag
		}(),
		commit: func() string {
			if len(commit) > 7 {
				return commit[:8]
			} else {
				return defaultCommit
			}
		}(),
	}

}

type buildOnAWSProvider struct {
	version string
	commit  string
}

func (p *buildOnAWSProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerName
	resp.Version = p.version
}

func (p *buildOnAWSProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: providerDesc,
		Attributes: map[string]tfsdk.Attribute{
			backendAddressField: {
				Description: backendAddressFieldDesc,
				Type:        types.StringType,
				Optional:    true,
			},
			skipTLSValidationField: {
				Description: skipTLSValidationFieldDesc,
				Type:        types.BoolType,
				Optional:    true,
			},
		},
	}, nil
}

func (p *buildOnAWSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	tflog.Info(ctx, "Configuring the BuildOnAWS provider")

	var config BuildOnAWSProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	backendAddressValue := backendAddressDefault

	if !config.BackendAddress.IsNull() {

		currentValue := config.BackendAddress.ValueString()
		tflog.Debug(ctx, "Backend URL set: "+currentValue)
		_, err := url.ParseRequestURI(currentValue)

		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root(backendAddressField),
				"Invalid URL for the backend address",
				"Cannot connect with the backend using the URL: '"+currentValue+"'.",
			)
			return
		} else {
			backendAddressValue = currentValue
		}

	}

	var insecureSkipVerify bool
	if !config.SkipTLSValidation.IsNull() && config.SkipTLSValidation.ValueBool() {
		insecureSkipVerify = true
		tflog.Debug(ctx, "NOT doing TLS certificate validation")
	}

	backendClient, _ := opensearch.NewClient(
		opensearch.Config{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
			},
			Addresses: []string{backendAddressValue},
		},
	)

	pingRequest := opensearchapi.PingRequest{
		Pretty:     true,
		Human:      true,
		ErrorTrace: true,
	}
	r, err := pingRequest.Do(ctx, backendClient)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failure connecting with the backend",
			"Reason: "+err.Error(),
		)
	} else {
		ctx = tflog.SetField(ctx, "ping_request_status", r.StatusCode)
		ctx = tflog.SetField(ctx, "ping_request_header", r.Header)
		ctx = tflog.SetField(ctx, "ping_request_body", r.Body)
		tflog.Debug(ctx, "Response from the ping request")
	}

	resp.DataSourceData = backendClient
	resp.ResourceData = backendClient

}

func (p *buildOnAWSProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCharacterDataSource,
	}
}

func (p *buildOnAWSProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCharacterResource,
	}
}
