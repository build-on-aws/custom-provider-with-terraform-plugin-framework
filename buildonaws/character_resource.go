package buildonaws

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

var (
	_ resource.Resource                = &characterResource{}
	_ resource.ResourceWithConfigure   = &characterResource{}
	_ resource.ResourceWithImportState = &characterResource{}
)

func NewCharacterResource() resource.Resource {
	return &characterResource{}
}

type characterResource struct {
	backendClient *opensearch.Client
}

func (r *characterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + resourceName
}

func (c *characterResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			idField: {
				Description: idFieldDesc,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			fullNameField: {
				Description: fullNameFieldDesc,
				Type:        types.StringType,
				Required:    true,
			},
			identityField: {
				Description: identityFieldDesc,
				Type:        types.StringType,
				Required:    true,
			},
			knownasField: {
				Description: knowasFieldDesc,
				Type:        types.StringType,
				Required:    true,
			},
			typeField: {
				Description: typeFieldDesc,
				Type:        types.StringType,
				Required:    true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf(characterTypes...),
				},
			},
			lastUpdatedField: {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (c *characterResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {

	tflog.Info(ctx, "Configuring the BuildOnAWS resource")

	if req.ProviderData == nil {
		return
	}

	c.backendClient = req.ProviderData.(*opensearch.Client)

}

func (c *characterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(idField), req, resp)
}

func (c *characterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var characterPlan CharacterResourceModel
	diags := req.Plan.Get(ctx, &characterPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	comicCharacter := &ComicCharacter{
		FullName: characterPlan.FullName.ValueString(),
		Identity: characterPlan.Identity.ValueString(),
		KnownAs:  characterPlan.KnownAs.ValueString(),
		Type:     characterPlan.Type.ValueString(),
	}
	bodyContent, _ := json.Marshal(comicCharacter)
	bodyReader := bytes.NewReader(bodyContent)
	indexRequest := opensearchapi.IndexRequest{
		Index: backendIndex,
		Body:  bodyReader,
	}

	indexResponse, err := indexRequest.Do(ctx, c.backendClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while creating character",
			"Reason: "+err.Error(),
		)
		return
	}

	defer indexResponse.Body.Close()
	bodyContent, err = io.ReadAll(indexResponse.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while reading response",
			"Reason: "+err.Error(),
		)
		return
	}

	backendResponse := &BackendResponse{}
	err = json.Unmarshal(bodyContent, backendResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while unmarshalling response",
			"Reason: "+err.Error(),
		)
		return
	}

	characterPlan.ID = types.StringValue(backendResponse.ID)
	characterPlan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, characterPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (c *characterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var characterState CharacterResourceModel
	diags := req.State.Get(ctx, &characterState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	documentID := characterState.ID.ValueString()

	getRequest := opensearchapi.GetRequest{
		Index:      backendIndex,
		DocumentID: documentID,
	}

	getResponse, err := getRequest.Do(ctx, c.backendClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while reading character",
			"Reason: "+err.Error(),
		)
		return
	}

	defer getResponse.Body.Close()
	bodyContent, err := io.ReadAll(getResponse.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while reading response",
			"Reason: "+err.Error(),
		)
		return
	}

	backendResponse := &BackendResponse{}
	err = json.Unmarshal(bodyContent, backendResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while unmarshalling response",
			"Reason: "+err.Error(),
		)
		return
	}

	characterState.FullName = types.StringValue(backendResponse.Source.FullName)
	characterState.Identity = types.StringValue(backendResponse.Source.Identity)
	characterState.KnownAs = types.StringValue(backendResponse.Source.KnownAs)
	characterState.Type = types.StringValue(backendResponse.Source.Type)

	diags = resp.State.Set(ctx, &characterState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (c *characterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var characterPlan CharacterResourceModel
	diags := req.Plan.Get(ctx, &characterPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	documentID := characterPlan.ID.ValueString()

	updateBody := &struct {
		Doc ComicCharacter `json:"doc,omitempty"`
	}{
		Doc: ComicCharacter{
			FullName: characterPlan.FullName.ValueString(),
			Identity: characterPlan.Identity.ValueString(),
			KnownAs:  characterPlan.KnownAs.ValueString(),
			Type:     characterPlan.Type.ValueString(),
		},
	}

	bodyContent, err := json.Marshal(updateBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while marshalling request",
			"Reason: "+err.Error(),
		)
		return
	}
	bodyReader := bytes.NewReader(bodyContent)
	updateRequest := opensearchapi.UpdateRequest{
		Index:      backendIndex,
		DocumentID: documentID,
		Body:       bodyReader,
	}

	_, err = updateRequest.Do(ctx, c.backendClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while updating character",
			"Reason: "+err.Error(),
		)
		return
	}

	characterPlan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, characterPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (c *characterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var characterState CharacterResourceModel
	diags := req.State.Get(ctx, &characterState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	documentID := characterState.ID.ValueString()
	deleteRequest := opensearchapi.DeleteRequest{
		Index:      backendIndex,
		DocumentID: documentID,
	}
	_, err := deleteRequest.Do(ctx, c.backendClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while deleting character",
			"Reason: "+err.Error(),
		)
		return
	}

}
