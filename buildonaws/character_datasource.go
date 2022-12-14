package buildonaws

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

var (
	_ datasource.DataSource              = &characterDataSource{}
	_ datasource.DataSourceWithConfigure = &characterDataSource{}
)

func NewCharacterDataSource() datasource.DataSource {
	return &characterDataSource{}
}

type characterDataSource struct {
	backendClient *opensearch.Client
}

func (c *characterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + characterDataSourceTypeName
}

func (c *characterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			idField: schema.StringAttribute{
				Description: idFieldDesc,
				Computed:    true,
			},
			fullNameField: schema.StringAttribute{
				Description: fullNameFieldDesc,
				Optional:    true,
				Computed:    true,
			},
			identityField: schema.StringAttribute{
				Description: identityFieldDesc,
				Required:    true,
			},
			knownasField: schema.StringAttribute{
				Description: knowasFieldDesc,
				Optional:    true,
				Computed:    true,
			},
			typeField: schema.StringAttribute{
				Description: typeFieldDesc,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (c *characterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {

	tflog.Info(ctx, "Configuring the BuildOnAWS datasource")

	if req.ProviderData == nil {
		return
	}

	c.backendClient = req.ProviderData.(*opensearch.Client)

}

func (c *characterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var characterPlan CharacterDataSourceModel
	diags := req.Config.Get(ctx, &characterPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	searchBody := &struct {
		Query struct {
			Match struct {
				Identity string `json:"identity,omitempty"`
			} `json:"match,omitempty"`
		} `json:"query,omitempty"`
	}{}

	searchBody.Query.Match.Identity = characterPlan.Identity.ValueString()
	bodyContent, _ := json.Marshal(searchBody)
	bodyReader := bytes.NewReader(bodyContent)
	searchRequest := opensearchapi.SearchRequest{
		Index: []string{backendIndex},
		Body:  bodyReader,
	}

	searchResponse, err := searchRequest.Do(ctx, c.backendClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while retrieving character",
			"Reason: "+err.Error(),
		)
	}
	defer searchResponse.Body.Close()

	bodyContent, err = io.ReadAll(searchResponse.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while reading response",
			"Reason: "+err.Error(),
		)
	}

	backendSearchResponse := &BackendSearchResponse{}
	err = json.Unmarshal(bodyContent, backendSearchResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while unmarshalling response",
			"Reason: "+err.Error(),
		)
		return
	}

	if backendSearchResponse.Hits.Total.Value > 0 {
		character := backendSearchResponse.Hits.Hits[0].Source
		characterPlan.ID = types.StringValue(character.ID)
		characterPlan.FullName = types.StringValue(character.FullName)
		characterPlan.Identity = types.StringValue(character.Identity)
		characterPlan.KnownAs = types.StringValue(character.KnownAs)
		characterPlan.Type = types.StringValue(character.Type)
	} else {
		var emptyString string
		characterPlan.ID = types.StringValue(emptyString)
		characterPlan.FullName = types.StringValue(emptyString)
		characterPlan.KnownAs = types.StringValue(emptyString)
		characterPlan.Type = types.StringValue(emptyString)
		resp.Diagnostics.AddWarning(
			"Datasource was not loaded",
			"Reason: no character with the identity '"+characterPlan.Identity.ValueString()+"'.",
		)
	}

	resp.State.Set(ctx, &characterPlan)
	diags = resp.State.Set(ctx, &characterPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}
