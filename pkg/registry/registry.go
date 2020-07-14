package registry

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

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
	credentials Credentials
	name        string

	registryClient RegistryClient
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
	return Registry{
		credentials:    c.Credentials,
		name:           c.Name,
		registryClient: c.RegistryClient,
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

func GetLink(linkHeader string) string {
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

func executeCmd(binary string, args []string) error {
	cmd := exec.Command(
		binary,
		args...,
	)

	var exitCode int = -1
	var exitErr *exec.ExitError

	output, err := cmd.CombinedOutput()
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}
	if err != nil {
		return microerror.Maskf(executionFailedError, "command execution failed with exit code = %d error = %#q and output:\n\n%s", exitCode, err, output)
	}

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
