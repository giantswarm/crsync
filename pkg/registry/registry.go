package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/giantswarm/microerror"
)

const (
	dockerBinaryName = "docker"
	tagLengthLimit   = 15
)

type Config struct {
	Credentials Credentials
	Name        string
	HttpClient  http.Client
}

type Registry struct {
	address     string
	credentials Credentials
	name        string
	httpClient  http.Client
}

type Credentials struct {
	User     string
	Password string
}

type Repository struct {
	Name string
	Tags []string
}

func New(c Config) (Registry, error) {

	// docker is specific with urls
	var registryAddress string
	{
		if c.Name == "docker.io" {
			registryAddress = "https://index.docker.io"
		} else {
			registryAddress = fmt.Sprintf("https://%s", c.Name)
		}
	}

	return Registry{
		address: registryAddress,
		credentials: Credentials{
			User:     c.Credentials.User,
			Password: c.Credentials.Password,
		},
		name:       c.Name,
		httpClient: c.HttpClient,
	}, nil

}

func (r *Registry) Login() error {
	fmt.Printf("Logging in destination container registry...\n")

	args := []string{"login", r.name, fmt.Sprintf("-u%s", r.credentials.User), fmt.Sprintf("-p%s", r.credentials.Password)}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) Logout() error {
	fmt.Printf("Logging out of destination container registry...\n")

	var args []string

	if r.name == "" {
		args = []string{"logout"}
	} else {
		args = []string{"logout", r.name}
	}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) ListRepositoryTags(repo string) ([]string, error) {
	fmt.Printf("\nReading list of tags from source registry for %#q repository...\n", repo)

	endpoint := fmt.Sprintf("%s/v2/%s/tags/list", r.address, repo)

	type tagsJSON struct {
		Tags []string `json:"tags"`
	}

	var tagsData tagsJSON

	var tags []string
	{
		nextEndpoint := fmt.Sprintf("%s", endpoint)
		for {
			resp, err := http.Get(nextEndpoint)
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
				nextEndpoint = fmt.Sprintf("%s%s", r.address, getLink(linkHeader))
			} else {
				break
			}
		}
	}

	return tags, nil

}

func (r *Registry) PullImage(repo, tag string) error {
	image := fmt.Sprintf("%s/%s:%s", r.name, repo, tag)

	args := []string{"pull", image}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) PushImage(repo, tag string) error {
	image := fmt.Sprintf("%s/%s:%s", r.name, repo, tag)

	args := []string{"push", image}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) RemoveImage(repo, tag string) error {
	image := fmt.Sprintf("%s/%s:%s", r.name, repo, tag)

	args := []string{"rmi", image}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func RetagImage(repo, tag, srcRegistry, dstRegistry string) error {

	srcImage := fmt.Sprintf("%s/%s:%s", srcRegistry, repo, tag)

	dstImage := fmt.Sprintf("%s/%s:%s", dstRegistry, repo, tag)

	fmt.Printf("Retagging image %#q into %#q\n", srcImage, dstImage)

	args := []string{"tag", srcImage, dstImage}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) RepositoryTagExists(repo, tag string) (bool, error) {
	endpoint := fmt.Sprintf("%s/v1/repositories/%s/tags/%s", r.address, repo, tag)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, microerror.Mask(err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if resp.StatusCode == 200 {
		return true, nil
	}

	return false, nil
}

func binaryExists() bool {
	cmd := exec.Command(dockerBinaryName)
	err := cmd.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.Sys().(syscall.WaitStatus).ExitStatus() == 0 {
				return true
			}
		}
		return false
	}
	return true
}

func executeCmd(binary string, args []string) error {
	if !binaryExists() {
		return microerror.Mask(executionFailedError)
	}

	cmd := exec.Command(
		binary,
		args...,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return microerror.Mask(err)
	}

	time.Sleep(time.Second * 1)

	return nil
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
