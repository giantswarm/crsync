package quay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/giantswarm/crsync/pkg/registry"
	"github.com/giantswarm/microerror"
)

const (
	publicImagesOnly   = true
	registryEndpoint   = "https://quay.io"
	repositoryEndpoint = "https://quay.io/api/v1/repository"
)

type Config struct {
	Namespace    string
	LastModified time.Duration
}

type Quay struct {
	namespace    string
	lastModified time.Duration

	httpClient *http.Client
}

func New(c Config) (*Quay, error) {
	httpClient := &http.Client{}

	if c.Namespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "Namespace must not be empty")
	}

	return &Quay{
		namespace:    c.Namespace,
		lastModified: c.LastModified,

		httpClient: httpClient,
	}, nil
}

func (q *Quay) Authorize() error {
	return nil
}

func (q *Quay) ListRepositories() ([]string, error) {
	fmt.Printf("Reading list of quay repostories in %#q namespace...\n", q.namespace)

	var reposToSync []string

	req, err := http.NewRequest("GET", repositoryEndpoint, nil)
	if err != nil {
		return reposToSync, microerror.Mask(err)
	}

	query := req.URL.Query()
	query.Add("last_modified", "true")
	query.Add("starred", "false")
	query.Add("public", fmt.Sprintf("%t", publicImagesOnly))
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
		lastModifiedTimestamp := time.Now().Add(-1 * q.lastModified).Unix()
		if int64(repo.LastModified) > lastModifiedTimestamp {
			reposToSync = append(reposToSync, fmt.Sprintf("%s/%s", q.namespace, repo.Name))
		}
	}

	return reposToSync, nil
}

func (q *Quay) ListTags(repository string) ([]string, error) {
	fmt.Printf("\nReading list of tags from source registry for %#q repository...\n", repository)

	endpoint := fmt.Sprintf("%s/v2/%s/tags/list", registryEndpoint, repository)

	type tagsJSON struct {
		Tags []string `json:"tags"`
	}

	var tagsData tagsJSON

	var tags []string
	{
		nextEndpoint := endpoint
		for {
			resp, err := http.Get(nextEndpoint) // #nosec G107
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
