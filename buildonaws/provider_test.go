package buildonaws

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		providerName: providerserver.NewProtocol6WithError(New()),
	}
)

func TestAccProviderAddressValidation(t *testing.T) {

	terrformConfig := `
	provider "buildonaws" {
		backend_address = "this-cannot-be-anything-you-want"
	}

	data "buildonaws_character" "deadpool" {
		identity = "Wade Wilson"
	}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      terrformConfig,
				ExpectError: regexp.MustCompile("Invalid URL for the backend address"),
			},
		},
	})

}

func TestAccProviderConnectivity(t *testing.T) {

	terrformConfig := `
	provider "buildonaws" {
		backend_address = "http://some-unknown-host:9200"
	}

	data "buildonaws_character" "deadpool" {
		identity = "Wade Wilson"
	}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      terrformConfig,
				ExpectError: regexp.MustCompile("Failure connecting with the backend"),
			},
		},
	})

}

func TestAccCharacterDataSource(t *testing.T) {

	ctx := context.Background()

	backendContainer, err := setupBackend(ctx)
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

func TestAccCharacterResource(t *testing.T) {

	ctx := context.Background()

	backendContainer, err := setupBackend(ctx)
	if err != nil {
		t.Error(err)
	}

	character := &ComicCharacter{
		FullName: "Daredevil",
		Identity: "Matt Murdock",
		KnownAs:  "The man without fear",
		Type:     characterTypes[1],
	}

	terraformConfig := `
	provider "buildonaws" {
		backend_address = "${backend_address}"
	}
	
	resource "buildonaws_character" "daredevil" {
		fullname = "${character_fullname}"
		identity = "${character_identity}"
		knownas = "${character_knownas}"
		type = "${character_type}"
	}`

	tfConfigCreateReadTest := terraformConfig
	tfConfigCreateReadTest = strings.ReplaceAll(tfConfigCreateReadTest, "${backend_address}", backendContainer.Address)
	tfConfigCreateReadTest = strings.ReplaceAll(tfConfigCreateReadTest, "${character_fullname}", character.FullName)
	tfConfigCreateReadTest = strings.ReplaceAll(tfConfigCreateReadTest, "${character_identity}", character.Identity)
	tfConfigCreateReadTest = strings.ReplaceAll(tfConfigCreateReadTest, "${character_knownas}", character.KnownAs)
	tfConfigCreateReadTest = strings.ReplaceAll(tfConfigCreateReadTest, "${character_type}", character.Type)

	tfConfigUpdateReadTest := terraformConfig
	dareDevilKnownAs := "The devil of hell's kitchen"
	tfConfigUpdateReadTest = strings.ReplaceAll(tfConfigUpdateReadTest, "${backend_address}", backendContainer.Address)
	tfConfigUpdateReadTest = strings.ReplaceAll(tfConfigUpdateReadTest, "${character_fullname}", character.FullName)
	tfConfigUpdateReadTest = strings.ReplaceAll(tfConfigUpdateReadTest, "${character_identity}", character.Identity)
	tfConfigUpdateReadTest = strings.ReplaceAll(tfConfigUpdateReadTest, "${character_knownas}", dareDevilKnownAs)
	tfConfigUpdateReadTest = strings.ReplaceAll(tfConfigUpdateReadTest, "${character_type}", character.Type)

	resourceName := "buildonaws_character.daredevil"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: tfConfigCreateReadTest,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify if the fields match with what is expected
					resource.TestCheckResourceAttr(resourceName, fullNameField, character.FullName),
					resource.TestCheckResourceAttr(resourceName, identityField, character.Identity),
					resource.TestCheckResourceAttr(resourceName, knownasField, character.KnownAs),
					resource.TestCheckResourceAttr(resourceName, typeField, character.Type),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet(resourceName, idField),
					resource.TestCheckResourceAttrSet(resourceName, lastUpdatedField),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{lastUpdatedField},
			},
			// Update and Read testing
			{
				Config: tfConfigUpdateReadTest,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify if the fields match with what is expected
					resource.TestCheckResourceAttr(resourceName, fullNameField, character.FullName),
					resource.TestCheckResourceAttr(resourceName, identityField, character.Identity),
					resource.TestCheckResourceAttr(resourceName, knownasField, dareDevilKnownAs),
					resource.TestCheckResourceAttr(resourceName, typeField, character.Type),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet(resourceName, idField),
					resource.TestCheckResourceAttrSet(resourceName, lastUpdatedField),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})

}

type backendContainer struct {
	Container testcontainers.Container
	Address   string
}

func setupBackend(ctx context.Context) (*backendContainer, error) {

	containerRequest := testcontainers.ContainerRequest{
		Image:        "opensearchproject/opensearch:2.4.0",
		Name:         "opensearch",
		ExposedPorts: []string{"9200/tcp"},
		Env: map[string]string{
			"cluster.name":                "opensearch-cluster",
			"node.name":                   "opensearch-node",
			"discovery.type":              "single-node",
			"bootstrap.memory_lock":       "true",
			"OPENSEARCH_JAVA_OPTS":        "-Xms1g -Xmx1g",
			"DISABLE_INSTALL_DEMO_CONFIG": "true",
			"DISABLE_SECURITY_PLUGIN":     "true",
		},
		WaitingFor: wait.ForLog("[opensearch-node] Node started"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerRequest,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "9200")
	if err != nil {
		return nil, err
	}

	return &backendContainer{
		Container: container,
		Address: fmt.Sprintf("http://%s:%s",
			host, mappedPort.Port())}, nil

}

func createCharacter(ctx context.Context, character *ComicCharacter,
	backendContainer *backendContainer) error {

	backendClient, err := opensearch.NewClient(
		opensearch.Config{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
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
