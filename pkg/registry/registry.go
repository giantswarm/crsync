package registry

import (
	"context"
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

func New(c Config) (*Registry, error) {
	return &Registry{
		name:           c.Name,
		registryClient: c.RegistryClient,
	}, nil
}

func (r *Registry) Login(ctx context.Context, user, password string) error {
	args := []string{"login", r.name, "-u", user, "-p", password}

	err := executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.authorize(ctx, user, password)
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

func (r *Registry) authorize(ctx context.Context, user, password string) error {
	return r.registryClient.Authorize(ctx, user, password)
}

func (r *Registry) ListRepositories(ctx context.Context) ([]string, error) {
	return r.registryClient.ListRepositories(ctx)
}

func (r *Registry) ListTags(ctx context.Context, repository string) ([]string, error) {
	return r.registryClient.ListTags(ctx, repository)
}

func (r Registry) Name() string {
	return r.name
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

	isRunning, err := imageIsRunning(repo, tag)
	if err != nil {
		return microerror.Mask(err)
	}
	if isRunning {
		return nil
	}

	args := []string{"rmi", image}

	err = executeCmd(dockerBinaryName, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
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

func imageIsRunning(repo, tag string) (bool, error) {
	cmd := exec.Command(dockerBinaryName, "ps")

	var exitCode int = -1
	var exitErr *exec.ExitError

	output, err := cmd.CombinedOutput()
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}
	if err != nil {
		return false, microerror.Maskf(executionFailedError, "command execution failed with exit code = %d error = %#q and output:\n\n%s", exitCode, err, output)
	}

	imageName := fmt.Sprintf("%s:%s", repo, tag)
	return strings.Contains(string(output), imageName), nil
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
