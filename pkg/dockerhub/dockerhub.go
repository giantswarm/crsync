package dockerhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
)

const (
	authEndpoint    = "https://hub.docker.com"
	registryAddress = "https://index.docker.io" // nolint
)

type Config struct {
}

type DockerHub struct {
	token string

	httpClient *http.Client
}

func New(c Config) (*DockerHub, error) {
	httpClient := &http.Client{}

	return &DockerHub{
		httpClient: httpClient,
	}, nil
}

func (d *DockerHub) Authorize(user, password string) error {

	endpoint := fmt.Sprintf("%s/v2/users/login/", authEndpoint)

	values := map[string]string{"username": user, "password": password}

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
	endpoint := fmt.Sprintf("%s/v2/repositories/%s/tags/", authEndpoint, repository)

	type dockerHubTags struct {
		Next    string `json:"next"`
		Results []struct {
			Name string `json:"name"`
		}
	}

	var tagsJSON dockerHubTags
	var tags []string
	{
		nextEndpoint := endpoint

		for {

			req, err := http.NewRequest("GET", nextEndpoint, nil)
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

			for _, tag := range tagsJSON.Results {
				tags = append(tags, tag.Name)
			}

			if tagsJSON.Next == nextEndpoint {
				break
			}

			nextEndpoint = tagsJSON.Next
		}
	}

	return tags, nil
}
