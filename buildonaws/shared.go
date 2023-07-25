package buildonaws

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/yaml.v2"
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

var (
	onlyOnce       sync.Once
	containerImage string
	containerName  string
	containerPort  string
	environment    map[string]string
)

func createBackendContainer(ctx context.Context) (*backendContainer, error) {

	onlyOnce.Do(func() {

		fileBytes, _ := os.ReadFile("../docker-compose.yml")
		dc := make(map[interface{}]interface{})
		yaml.Unmarshal(fileBytes, &dc)

		containerImage = dc["services"].(map[interface{}]interface{})["opensearch"].(map[interface{}]interface{})["image"].(string)
		containerName = dc["services"].(map[interface{}]interface{})["opensearch"].(map[interface{}]interface{})["container_name"].(string)
		cp := dc["services"].(map[interface{}]interface{})["opensearch"].(map[interface{}]interface{})["ports"].([]interface{})[0]
		containerPort = strings.Split(cp.(string), ":")[0]

		environment = make(map[string]string)
		env := dc["services"].(map[interface{}]interface{})["opensearch"].(map[interface{}]interface{})["environment"].([]interface{})
		for _, item := range env {
			entryParts := strings.Split(item.(string), "=")
			environment[entryParts[0]] = entryParts[1]
		}

	})

	if containerImage == "" || containerName == "" || containerPort == "" || environment == nil {
		return nil, errors.New("something wrong with the Docker Compose file")
	}

	containerRequest := testcontainers.ContainerRequest{
		Image:        containerImage,
		Name:         containerName,
		ExposedPorts: []string{containerPort + "/tcp"},
		Env:          environment,
		WaitingFor:   wait.ForLog("[opensearch-node] Node started"),
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

	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return nil, err
	}

	return &backendContainer{
		Container: container,
		Address: fmt.Sprintf("http://%s:%s",
			host, mappedPort.Port())}, nil

}
