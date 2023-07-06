package buildonaws

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCharacterResource(t *testing.T) {

	ctx := context.Background()

	backendContainer, err := createBackendContainer(ctx)
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
			// ImportState testing
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
