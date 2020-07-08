package quay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/giantswarm/microerror"
)

const (
	publicImagesOnly   = true
	repositoryEndpoint = "https://quay.io/api/v1/repository"
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
