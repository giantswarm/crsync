package dockerhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/crsync/pkg/registry"
)

const (
	authEndpoint    = "https://hub.docker.com"
	registryAddress = "https://index.docker.io" // nolint
)

type Config struct {
	Credentials registry.Credentials
}

type DockerHub struct {
	token       string
	credentials registry.Credentials

	httpClient *http.Client
}

func New(c Config) (*DockerHub, error) {
	if c.Credentials.User == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.User must not be empty", c)
	}
	if c.Credentials.Password == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Password must not be empty", c)
	}

	httpClient := &http.Client{}

	return &DockerHub{
		credentials: c.Credentials,

		httpClient: httpClient,
	}, nil
}

func (d *DockerHub) Authorize() error {

	endpoint := fmt.Sprintf("%s/v2/users/login/", authEndpoint)

	values := map[string]string{"username": d.credentials.User, "password": d.credentials.Password}

	jsonValues, _ := json.Marshal(values)

	resp, err := d.httpClient.Post(endpoint, "application/json", bytes.NewBuffer(jsonValues)) // nolint
	if err != nil {
		return microerror.Mask(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return microerror.Mask(err)
	}

	type authDataResponse struct {
		Token string `json:"token"`
	}

	var authData authDataResponse
	err = json.Unmarshal(body, &authData)
	if err != nil {
		return microerror.Mask(err)
	}

	d.token = authData.Token

	return nil
}

func (d *DockerHub) ListRepositories() ([]string, error) {
	return nil, microerror.Maskf(executionFailedError, "method not implemented")
}

func (d *DockerHub) ListTags(repository string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/v2/repositories/%s/tags/?page_size=10000", authEndpoint, repository)

	type dockerHubTags struct {
		Results []struct {
			Name string `yaml:"name"`
		}
	}

	var tagsJSON dockerHubTags
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("JWT %s", d.token))

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	err = json.Unmarshal(body, &tagsJSON)
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	var tags []string
	{
		for _, tag := range tagsJSON.Results {
			tags = append(tags, tag.Name)
		}
	}

	return tags, nil
}
