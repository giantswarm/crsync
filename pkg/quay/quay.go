package quay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/crsync/internal/env"
	"github.com/giantswarm/crsync/pkg/registry"
)

const (
	registryEndpoint   = "https://quay.io"
	repositoryEndpoint = "https://quay.io/api/v1/repository"
)

type Config struct {
	Namespace               string
	LastModified            time.Duration
	SyncPrivateRepositories bool
}

type Quay struct {
	namespace               string
	lastModified            time.Duration
	token                   string
	syncPrivateRepositories bool

	httpClient *http.Client
}

func New(c Config) (*Quay, error) {
	httpClient := &http.Client{}

	if c.Namespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", c)
	}

	quayToken := os.Getenv(env.QuayApiToken)

	if quayToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "Environment variable %#q must contain valid quay token", env.QuayApiToken)
	}

	return &Quay{
		namespace:               c.Namespace,
		lastModified:            c.LastModified,
		token:                   quayToken,
		syncPrivateRepositories: c.SyncPrivateRepositories,

		httpClient: httpClient,
	}, nil
}

func (q *Quay) Authorize(user, password string) error {
	return nil
}

func (q *Quay) ListRepositories() ([]string, error) {
	var reposToSync []string

	req, err := http.NewRequest("GET", repositoryEndpoint, nil)
	if err != nil {
		return reposToSync, microerror.Mask(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", q.token))

	query := req.URL.Query()
	query.Add("last_modified", "true")
	query.Add("starred", "false")
	query.Add("namespace", q.namespace)
	req.URL.RawQuery = query.Encode()

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return reposToSync, microerror.Mask(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return reposToSync, microerror.Mask(err)
	}

	var data RepositoriesJSON

	err = json.Unmarshal(body, &data)
	if err != nil {
		return reposToSync, microerror.Mask(err)
	}

	for _, repo := range data.Repositories {
		if !repo.IsPublic && !q.syncPrivateRepositories {
			continue
		}

		lastModifiedTimestamp := time.Now().Add(-1 * q.lastModified).Unix()
		if int64(repo.LastModified) > lastModifiedTimestamp {
			reposToSync = append(reposToSync, fmt.Sprintf("%s/%s", q.namespace, repo.Name))
		}
	}

	return reposToSync, nil
}

func (q *Quay) ListTags(repository string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/v2/%s/tags/list", registryEndpoint, repository)

	type tagsJSON struct {
		Tags []string `json:"tags"`
	}

	var tagsData tagsJSON

	var tags []string
	{
		nextEndpoint := endpoint
		for {
			req, err := http.NewRequest("GET", nextEndpoint, nil)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", q.token))

			resp, err := q.httpClient.Do(req)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			err = json.Unmarshal(body, &tagsData)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			tags = append(tags, tagsData.Tags...)

			linkHeader := resp.Header.Get("Link")
			if linkHeader == "" {
				break
			}

			nextEndpoint = fmt.Sprintf("%s%s", registryEndpoint, registry.GetLink(linkHeader))
		}
	}

	return tags, nil

}
