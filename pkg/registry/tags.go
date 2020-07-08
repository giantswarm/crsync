package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
)

func listRepoTagsAzureCR(endpoint, token, repo string, httpClient *http.Client) ([]string, error) {
	endpoint = fmt.Sprintf("%s/v2/%s/tags/list", endpoint, repo)

	type azureCRTags struct {
		Tags []string `yaml:"tags"`
	}

	var tagsJSON azureCRTags
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("basic %s", token))

	resp, err := httpClient.Do(req)
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

	return tagsJSON.Tags, nil
}
