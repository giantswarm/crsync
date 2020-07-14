package registry

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/giantswarm/microerror"
)

const (
	dockerBinaryName = "docker"
)

type Config struct {
	Name           string
	HttpClient     http.Client
	RegistryClient RegistryClient
}

type Registry struct {
	name string

	registryClient RegistryClient
}

type Repository struct {
	Name string
	Tags []string
}

func New(c Config) (Registry, error) {
	return Registry{
		name:           c.Name,
		registryClient: c.RegistryClient,
	}, nil

}

func (r *Registry) Login(user, password string) error {
	fmt.Printf("Logging in destination container registry...\n")

	args := []string{"login", r.name, "-u", user, "-p", password}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.authorize(user, password)
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

func (r *Registry) authorize(user, password string) error {
	return r.registryClient.Authorize(user, password)
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
