package buildonaws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		providerTypeName: providerserver.NewProtocol6WithError(New()),
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
