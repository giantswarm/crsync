package registry

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/giantswarm/microerror"
)

const (
	dockerBinaryName = "docker"
)

type Config struct {
	Credentials    Credentials
	Name           string
	HttpClient     http.Client
	RegistryClient RegistryClient
}

type Registry struct {
	address     string
	auth        Auth
	credentials Credentials
	name        string
	kind        string

	registryClient RegistryClient
	httpClient     *http.Client
}
type Auth struct {
	endpoint string
	token    string
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
	var registryAddress, authEndpoint, kind string
	{
		if c.Name == "docker.io" {
			registryAddress = "https://index.docker.io"
			authEndpoint = "https://hub.docker.com"
			kind = DockerHubContainerRegistry
		} else {
			registryAddress = fmt.Sprintf("https://%s", c.Name)
			authEndpoint = fmt.Sprintf("https://%s", c.Name)
			kind = AzureContainerRegistry
		}
	}

	return Registry{
		address: registryAddress,
		auth: Auth{
			endpoint: authEndpoint,
		},
		credentials:    Credentials(c.Credentials),
		kind:           kind,
		name:           c.Name,
		registryClient: c.RegistryClient,
		httpClient:     &c.HttpClient,
	}, nil

}

func (r *Registry) Login() error {
	fmt.Printf("Logging in destination container registry...\n")

	args := []string{"login", r.name, fmt.Sprintf("-u%s", r.credentials.User), fmt.Sprintf("-p%s", r.credentials.Password)}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.authorize()
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

func (r *Registry) authorize() error {
	return r.registryClient.Authorize()
}

func (r *Registry) ListRepositories() ([]string, error) {
	return r.registryClient.ListRepositories()
}

func (r *Registry) ListTags(repository string) ([]string, error) {
	return r.registryClient.ListTags(repository)
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
	var tags []string
	var err error

	tags, err = r.ListTags(repo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return stringInSlice(tag, tags), nil
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
