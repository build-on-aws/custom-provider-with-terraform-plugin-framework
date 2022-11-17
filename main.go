package main

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name buildonaws

func main() {

	var debug bool
	flag.BoolVar(&debug, "debug", false, "set this to true if you want to debug the code using delve")
	flag.Parse()

	ctx := context.Background()

	providerserver.Serve(ctx, NewBuildOnAWSProvider, providerserver.ServeOpts{
		Debug:   debug,
		Address: "aws.amazon.com/terraform/buildonaws",
	})

}
