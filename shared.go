package main

import "strings"

var (
	providerName            = "buildonaws"
	backendIndex            = providerName
	providerDesc            = "Provider to manage characters from comic books."
	backendAddressField     = "backend_address"
	backendAddressFieldDesc = "Address to connect to the OpenSearch backend."
	backendAddressDefault   = "http://localhost:9200"
)

var (
	dataSourceName    = "_character"
	resourceName      = dataSourceName
	idField           = "id"
	idFieldDesc       = "Unique identifier of the character."
	fullNameField     = "fullname"
	fullNameFieldDesc = "The name to which we know the character of."
	identityField     = "identity"
	identityFieldDesc = "The real name of the character, which is usually a secret."
	knownasField      = "knownas"
	knowasFieldDesc   = "A catchphrase for which we know the character of."
	typeField         = "type"
	characterTypes    = []string{"hero", "super-hero", "anti-hero", "villain"}
	typeFieldDesc     = "The type of character. Possible values: '" + strings.Join(characterTypes, ",") + "'."
	lastUpdatedField  = "last_updated"
)
