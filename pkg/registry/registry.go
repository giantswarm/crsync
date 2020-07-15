package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

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
	dockerClient   *client.Client
}

type Repository struct {
	Name string
	Tags []string
}

func New(c Config) (*Registry, error) {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &Registry{
		name:           c.Name,
		registryClient: c.RegistryClient,
		dockerClient:   dockerClient,
	}, nil
}

func (r *Registry) Login(ctx context.Context, user, password string) error {
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

func (r *Registry) Logout(ctx context.Context) error {
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

func (r *Registry) ListRepositories(ctx context.Context) ([]string, error) {
	return r.registryClient.ListRepositories()
}

func (r *Registry) ListTags(ctx context.Context, repository string) ([]string, error) {
	return r.registryClient.ListTags(repository)
}

func (r *Registry) Pull(ctx context.Context, repo, tag string) error {
	image := fmt.Sprintf("%s/%s:%s", r.name, repo, tag)

	args := []string{"pull", image}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) Push(ctx context.Context, repo, tag string) error {
	image := fmt.Sprintf("%s/%s:%s", r.name, repo, tag)

	args := []string{"push", image}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) RemoveImage(ctx context.Context, repo, tag string) error {
	image := fmt.Sprintf("%s/%s:%s", r.name, repo, tag)

	runningImages, err := r.listRunningImages()
	if err != nil {
		return microerror.Mask(err)
	}
	if _, ok := runningImages[image]; ok {
		// Host is currently running this particular image; Trying to remove it
		// would cause errors.
		return nil
	}

	args := []string{"rmi", image}

	err = executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Registry) listRunningImages() (map[string]bool, error) {
	images := map[string]bool{}

	containers, err := r.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return images, microerror.Mask(err)
	}

	for _, c := range containers {
		images[c.Image] = true
	}

	return images, nil
}

func RetagImage(repo, tag, srcRegistry, dstRegistry string) error {
	srcImage := fmt.Sprintf("%s/%s:%s", srcRegistry, repo, tag)
	dstImage := fmt.Sprintf("%s/%s:%s", dstRegistry, repo, tag)

	args := []string{"tag", srcImage, dstImage}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
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
