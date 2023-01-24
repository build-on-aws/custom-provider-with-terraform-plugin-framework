package buildonaws

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	providerTypeName        = "buildonaws"
	backendIndex            = providerTypeName
	providerDesc            = "Provider to manage characters from comic books."
	backendAddressField     = "backend_address"
	backendAddressFieldDesc = "Address to connect to the OpenSearch backend."
	backendAddressDefault   = "http://localhost:9200"
)

var (
	characterDataSourceTypeName = "_character"
	characterResourceTypeName   = characterDataSourceTypeName
	idField                     = "id"
	idFieldDesc                 = "Unique identifier of the character."
	fullNameField               = "fullname"
	fullNameFieldDesc           = "The name to which we know the character of."
	identityField               = "identity"
	identityFieldDesc           = "The real name of the character, which is usually a secret."
	knownasField                = "knownas"
	knowasFieldDesc             = "A catchphrase for which we know the character of."
	typeField                   = "type"
	characterTypes              = []string{"hero", "super-hero", "anti-hero", "villain"}
	typeFieldDesc               = "The type of character. Possible values: '" + strings.Join(characterTypes, ",") + "'."
	lastUpdatedField            = "last_updated"
)

type backendContainer struct {
	Container testcontainers.Container
	Address   string
}

func setupBackend(ctx context.Context) (*backendContainer, error) {

	containerRequest := testcontainers.ContainerRequest{
		Image:        "opensearchproject/opensearch:2.5.0",
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
