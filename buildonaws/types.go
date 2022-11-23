package buildonaws

import "github.com/hashicorp/terraform-plugin-framework/types"

type BuildOnAWSProviderModel struct {
	BackendAddress    types.String `tfsdk:"backend_address"`
	SkipTLSValidation types.Bool   `tfsdk:"skip_tls_validation"`
}

type CharacterDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	FullName types.String `tfsdk:"fullname"`
	Identity types.String `tfsdk:"identity"`
	KnownAs  types.String `tfsdk:"knownas"`
	Type     types.String `tfsdk:"type"`
}

type CharacterResourceModel struct {
	ID          types.String `tfsdk:"id"`
	FullName    types.String `tfsdk:"fullname"`
	Identity    types.String `tfsdk:"identity"`
	KnownAs     types.String `tfsdk:"knownas"`
	Type        types.String `tfsdk:"type"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

type ComicCharacter struct {
	ID       string `json:"_id,omitempty"`
	FullName string `json:"fullname,omitempty"`
	Identity string `json:"identity,omitempty"`
	KnownAs  string `json:"knownas,omitempty"`
	Type     string `json:"type,omitempty"`
}

type BackendResponse struct {
	ID     string          `json:"_id"`
	Source *ComicCharacter `json:"_source"`
}

type BackendSearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []*struct {
			ID     string          `json:"_id"`
			Source *ComicCharacter `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
