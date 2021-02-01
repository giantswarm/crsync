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
	tagsPerPage     = 10
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

func (d *DockerHub) CountTags(repository string) (int, error) {
	return 0, microerror.Maskf(executionFailedError, "method not implemented")
}

func (d *DockerHub) ListRepositories() ([]string, error) {
	return nil, microerror.Maskf(executionFailedError, "method not implemented")
}

func (d *DockerHub) ListTags(repository string, limit int) ([]string, error) {
	page := 1
	endpoint := fmt.Sprintf("%s/v2/repositories/%s/tags/?page=%d", authEndpoint, repository, page)

	type dockerHubTags struct {
		Count   int    `json:"count"`
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

			if resp.StatusCode != http.StatusOK {
				return []string{}, microerror.Maskf(executionFailedError, fmt.Sprintf("Expected status code '%d', got '%d': %v", http.StatusOK, resp.StatusCode, string(body)))
			}

			err = json.Unmarshal(body, &tagsJSON)
			if err != nil {
				return []string{}, microerror.Mask(err)
			}

			fmt.Printf("\nRepo: %s | Page: %d\n", repository, page)

			for i, tag := range tagsJSON.Results {
				fmt.Printf("%d: tag: %s\n", i+1, tag.Name)
				tags = append(tags, tag.Name)
			}

			fmt.Println()

			numOfPages := (tagsJSON.Count / tagsPerPage) + 1
			if page == numOfPages {
				break
			}

			nextEndpoint = tagsJSON.Next
			page += 1
		}
	}

	return tags, nil
}
