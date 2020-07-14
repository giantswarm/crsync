package azurecr

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/crsync/pkg/registry"
)

type Config struct {
	RegistryName string
}

type AzureCR struct {
	token            string
	registryEndpoint string

	httpClient *http.Client
}

func New(c Config) (*AzureCR, error) {
	if c.RegistryName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryName must not be empty", c)
	}

	httpClient := &http.Client{}

	return &AzureCR{
		registryEndpoint: fmt.Sprintf("https://%s", c.RegistryName),

		httpClient: httpClient,
	}, nil
}

func (d *AzureCR) Authorize(user, password string) error {

	b64creds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, password)))

	d.token = b64creds

	return nil
}

func (d *AzureCR) ListRepositories() ([]string, error) {
	return nil, microerror.Maskf(executionFailedError, "method not implemented")
}

func (d *AzureCR) ListTags(repository string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/v2/%s/tags/list", d.registryEndpoint, repository)

	type azureCRTags struct {
		Tags []string `yaml:"tags"`
	}

	var tagsJSON azureCRTags
	var tags []string
	{
		nextEndpoint := endpoint

		for {
			req, err := http.NewRequest("GET", nextEndpoint, nil)
			if err != nil {
				return []string{}, microerror.Mask(err)
			}

			req.Header.Set("Authorization", fmt.Sprintf("basic %s", d.token))

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
			tags = append(tags, tagsJSON.Tags...)

			linkHeader := resp.Header.Get("Link")
			if linkHeader == "" {
				break
			}

			nextEndpoint = fmt.Sprintf("%s%s", d.registryEndpoint, registry.GetLink(linkHeader))
		}
	}

	return tags, nil
}
