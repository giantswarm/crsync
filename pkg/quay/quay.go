package quay

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/giantswarm/microerror"
)

const (
	registryEndpoint   = "https://quay.io"
	repositoryEndpoint = "https://quay.io/api/v1/repository"
)

type Config struct {
	Namespace                  string
	LastModified               time.Duration
	Token                      string
	IncludePrivateRepositories bool
}

type Quay struct {
	namespace                  string
	lastModified               time.Duration
	token                      string
	includePrivateRepositories bool

	httpClient *http.Client
}

func New(c Config) (*Quay, error) {
	httpClient := &http.Client{}

	if c.Namespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", c)
	}
	if c.IncludePrivateRepositories && c.Token == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Token must not be empty", c)
	}

	return &Quay{
		namespace:                  c.Namespace,
		lastModified:               c.LastModified,
		token:                      c.Token,
		includePrivateRepositories: c.IncludePrivateRepositories,

		httpClient: httpClient,
	}, nil
}

func (q *Quay) Authorize(ctx context.Context, user, password string) error {
	return nil
}

func (q *Quay) ListRepositories(ctx context.Context) ([]string, error) {
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
		if !repo.IsPublic && !q.includePrivateRepositories {
			continue
		}

		lastModifiedTimestamp := time.Now().Add(-1 * q.lastModified).Unix()
		if int64(repo.LastModified) > lastModifiedTimestamp {
			reposToSync = append(reposToSync, fmt.Sprintf("%s/%s", q.namespace, repo.Name))
		}
	}

	return reposToSync, nil
}

func (q *Quay) ListTags(ctx context.Context, repository string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/api/v1/repository/%s/tag/", registryEndpoint, repository)

	type dataJSON struct {
		HasAdditional bool `json:"has_additional"`
		Tags          []struct {
			Name string `json:"name"`
		} `json:"tags"`
	}

	var tagsData dataJSON

	var tags []string
	{
		page := 1
		hasAdditional := true
		for hasAdditional {
			req, err := http.NewRequest("GET", endpoint, nil)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", q.token))

			query := req.URL.Query()
			query.Add("page", strconv.Itoa(page))
			query.Add("onlyActiveTags", fmt.Sprintf("%t", true))

			req.URL.RawQuery = query.Encode()

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

			for _, tag := range tagsData.Tags {
				tags = append(tags, tag.Name)
			}

			hasAdditional = tagsData.HasAdditional
			page++
		}
	}

	return tags, nil

}
