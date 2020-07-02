package quayio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
)

const (
	// Sun, 01 Jul 18 00:00:00 +0000
	lastModifiedTimestamp = 1530403200

	publicImagesOnly   = true
	repositoryEndpoint = "https://quay.io/api/v1/repository"
)

type Repository struct {
	Name         string `json:"name"`
	LastModified int    `json:"last_modified"`
}

type RepositoriesJSON struct {
	Repositories []Repository `json:"repositories"`
}

func ListRepositories(namespace string) ([]string, error) {
	fmt.Printf("Reading list of quay repostories in %#q namespace...\n", namespace)

	var reposToSync []string

	httpClient := http.Client{}

	req, err := http.NewRequest("GET", repositoryEndpoint, nil)
	if err != nil {
		return reposToSync, microerror.Mask(err)
	}

	q := req.URL.Query()
	q.Add("last_modified", "true")
	q.Add("starred", "false")
	q.Add("public", fmt.Sprintf("%t", publicImagesOnly))
	q.Add("namespace", namespace)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
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
		if repo.LastModified > lastModifiedTimestamp {
			reposToSync = append(reposToSync, fmt.Sprintf("%s/%s", namespace, repo.Name))
		}
	}

	return reposToSync, nil
}
