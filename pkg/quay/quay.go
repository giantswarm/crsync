package quay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
)

const (
	publicImagesOnly   = true
	registryEndpoint   = "https://quay.io"
	repositoryEndpoint = "https://quay.io/api/v1/repository"

	tagLengthLimit = 20
)

type Quay struct {
	namespace    string
	lastModified time.Duration

	httpClient *http.Client
}

type Config struct {
	Namespace    string
	LastModified time.Duration
}

type Repository struct {
	Name         string `json:"name"`
	LastModified int    `json:"last_modified"`
}

type RepositoriesJSON struct {
	Repositories []Repository `json:"repositories"`
}

func New(c Config) (*Quay, error) {
	httpClient := &http.Client{}

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
		lastModifiedTimestamp := int(time.Now().Add(-1 * q.lastModified).Unix())
		if repo.LastModified > lastModifiedTimestamp {
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
			resp, err := http.Get(nextEndpoint) // nolint
			if err != nil {
				return []string{}, microerror.Mask(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return []string{}, microerror.Mask(err)
			}

			err = json.Unmarshal(body, &tagsData)
			if err != nil {
				return []string{}, microerror.Mask(err)
			}

			for _, tag := range tagsData.Tags {
				if len(tag) < tagLengthLimit {
					tags = append(tags, tag)
				}
			}

			linkHeader := resp.Header.Get("Link")
			if linkHeader != "" {
				nextEndpoint = fmt.Sprintf("%s%s", registryEndpoint, getLink(linkHeader))
			} else {
				break
			}
		}
	}

	return tags, nil

}

func getLink(linkHeader string) string {
	start := "<"
	end := ">"
	s := strings.Index(linkHeader, start)
	if s == -1 {
		return ""
	}
	s += len(start)
	e := strings.Index(linkHeader, end)
	if e == -1 {
		return ""
	}
	return linkHeader[s:e]
}
