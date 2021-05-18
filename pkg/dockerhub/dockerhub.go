package dockerhub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/types"
	"github.com/giantswarm/microerror"
)

const (
	authEndpoint    = "https://hub.docker.com"
	registryAddress = "https://index.docker.io" // nolint
)

type Config struct {
}

type DockerHub struct {
	user     string
	password string
	token    string

	httpClient *http.Client
}

func New(c Config) (*DockerHub, error) {
	httpClient := &http.Client{}

	return &DockerHub{
		httpClient: httpClient,
	}, nil
}

func (d *DockerHub) Authorize(ctx context.Context, user, password string) error {

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

	d.user = user
	d.password = password
	d.token = authData.Token

	return nil
}

func (d *DockerHub) ListRepositories(ctx context.Context) ([]string, error) {
	return nil, microerror.Maskf(executionFailedError, "method not implemented")
}

func (d *DockerHub) ListTags(ctx context.Context, repository string) ([]string, error) {
	if d.user == "" || d.password == "" {
		return nil, microerror.Maskf(executionFailedError, "can not run ListTags without calling Authorize first")
	}

	sys := &types.SystemContext{
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: d.user,
			Password: d.password,
		},
	}

	ref, err := reference.ParseNormalizedNamed(repository)
	if err != nil {
		return nil, microerror.Maskf(executionFailedError, "failed to parse repository %#q with error: %s", repository, err)
	}

	taggedRef, err := docker.NewReference(reference.TagNameOnly(ref))
	if err != nil {
		return nil, microerror.Maskf(executionFailedError, "failed to convert ref %#q to a tagged one with error: %s", ref, err)
	}

	tags, err := docker.GetRepositoryTags(ctx, sys, taggedRef)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "404 (not found)") {
		// Docker registry returns 404 when there are no tags for the
		// repository.
		tags = []string{}
	} else if err != nil {
		return nil, microerror.Maskf(executionFailedError, "failed to get tags for ref %#q with error: %s", ref, err)
	}

	return tags, nil
}
