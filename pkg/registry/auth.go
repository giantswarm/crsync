package registry

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
)

const (
	AzureContainerRegistry     = "azurecr"
	DockerHubContainerRegistry = "dockerhub"
)

var (
	SupportedAuthMethods = []string{AzureContainerRegistry, DockerHubContainerRegistry}
)

func (r *Registry) Authorize() error {
	fmt.Printf("\nAuthorizing in registry %#q...\n", r.name)

	var token string
	var err error

	switch r.kind {
	case DockerHubContainerRegistry:
		token, err = dockerHubAuth(r.auth.endpoint, r.credentials)
		if err != nil {
			return microerror.Mask(err)
		}
	case AzureContainerRegistry:
		token, err = basicAuth(r.credentials)
		if err != nil {
			return microerror.Mask(err)
		}
	default:
		return false, microerror.Maskf(executionFailedError, "uknonw container registry kind %#q", r.kind)
	}

	r.auth.token = token

	return nil
}

func IsSupportedAuthMethod(authMethod string) bool {
	for _, m := range SupportedAuthMethods {
		if m == authMethod {
			return true
		}
	}
	return false
}

func basicAuth(creds Credentials) (string, error) {
	b64creds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", creds.User, creds.Password)))

	return b64creds, nil
}

func dockerHubAuth(authEndpoint string, creds Credentials) (string, error) {

	endpoint := fmt.Sprintf("%s/v2/users/login/", authEndpoint)

	values := map[string]string{"username": creds.User, "password": creds.Password}

	jsonValues, _ := json.Marshal(values)

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValues)) // nolint
	if err != nil {
		return "", microerror.Mask(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", microerror.Mask(err)
	}

	type authDataResponse struct {
		Token string `json:"token"`
	}

	var authData authDataResponse
	err = json.Unmarshal(body, &authData)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return authData.Token, nil
}
