package buildonaws

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

func TestAccCharacterDataSource(t *testing.T) {

	ctx := context.Background()

	backendContainer, err := createBackendContainer(ctx)
	if err != nil {
		t.Error(err)
	}

	character := &ComicCharacter{
		FullName: "Deadpool",
		Identity: "Wade Wilson",
		KnownAs:  "Merch with a mouth",
		Type:     characterTypes[2],
	}

	err = createCharacter(ctx, character, backendContainer)
	if err != nil {
		t.Error(err)
	}

	terrformConfig := `
	provider "buildonaws" {
		backend_address = "${backend_address}"
	}
	
	data "buildonaws_character" "deadpool" {
		identity = "${character_identity}"
	}`

	terrformConfig = strings.ReplaceAll(terrformConfig, "${backend_address}", backendContainer.Address)
	terrformConfig = strings.ReplaceAll(terrformConfig, "${character_identity}", character.Identity)

	dataSourceName := "data.buildonaws_character.deadpool"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: terrformConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify if the fields match with what is expected
					resource.TestCheckResourceAttr(dataSourceName, fullNameField, character.FullName),
					resource.TestCheckResourceAttr(dataSourceName, knownasField, character.KnownAs),
					resource.TestCheckResourceAttr(dataSourceName, typeField, character.Type),
				),
			},
		},
	})

}

func createCharacter(ctx context.Context, character *ComicCharacter,
	backendContainer *backendContainer) error {

	backendClient, err := opensearch.NewClient(
		opensearch.Config{
			Addresses: []string{backendContainer.Address},
		},
	)
	if err != nil {
		return err
	}

	bodyContent, err := json.Marshal(character)
	if err != nil {
		return err
	}

	bodyReader := bytes.NewReader(bodyContent)
	indexRequest := opensearchapi.IndexRequest{
		Index:   backendIndex,
		Body:    bodyReader,
		Refresh: "wait_for",
	}

	_, err = indexRequest.Do(ctx, backendClient)
	if err != nil {
		return err
	}

	return nil

}
