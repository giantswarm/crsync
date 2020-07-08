package registry

import (
	"encoding/base64"
	"fmt"
)

const (
	AzureContainerRegistry     = "azurecr"
	DockerHubContainerRegistry = "dockerhub"
)

var (
	SupportedAuthMethods = []string{AzureContainerRegistry, DockerHubContainerRegistry}
)

func basicAuth(creds Credentials) (string, error) {
	b64creds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", creds.User, creds.Password)))

	return b64creds, nil
}
