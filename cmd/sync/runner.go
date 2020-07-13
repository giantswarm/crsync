package sync

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/crsync/internal/key"
	"github.com/giantswarm/crsync/pkg/azurecr"
	"github.com/giantswarm/crsync/pkg/dockerhub"
	"github.com/giantswarm/crsync/pkg/quay"
	"github.com/giantswarm/crsync/pkg/registry"
)

const (
	sourceRegistryName = "quay.io"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	var srcRegistryClient registry.RegistryClient
	{
		c := quay.Config{
			Namespace:    key.Namespace,
			LastModified: r.flag.LastModified,
		}

		srcRegistryClient, err = quay.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var srcRegistry registry.Registry
	{
		c := registry.Config{
			Name:           sourceRegistryName,
			RegistryClient: srcRegistryClient,
		}

		srcRegistry, err = registry.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var dstRegistryClient registry.RegistryClient
	{
		switch registryName := r.flag.DstRegistryName; {
		case registryName == "docker.io":
			c := dockerhub.Config{}

			dstRegistryClient, err = dockerhub.New(c)
			if err != nil {
				return microerror.Mask(err)
			}
		case strings.HasSuffix(registryName, "azurecr.io"):
			c := azurecr.Config{
				RegistryName: r.flag.DstRegistryName,
			}

			dstRegistryClient, err = azurecr.New(c)
			if err != nil {
				return microerror.Mask(err)
			}

		default:
			return microerror.Maskf(executionFailedError, "unknown container registry %#q", registryName)
		}
	}

	var dstRegistry registry.Registry
	{
		config := registry.Config{
			Name:           r.flag.DstRegistryName,
			RegistryClient: dstRegistryClient,
		}

		dstRegistry, err = registry.New(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	defer func() { _ = dstRegistry.Logout() }()
	err = dstRegistry.Login(r.flag.DstRegistryUser, r.flag.DstRegistryPassword)
	if err != nil {
		return microerror.Mask(err)
	}

	reposToSync, err := srcRegistry.ListRepositories()
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("There are %d repositories to sync.\n", len(reposToSync))

	for repoIndex, repo := range reposToSync {
		tags, err := srcRegistry.ListTags(repo)
		if err != nil {
			return microerror.Mask(err)
		}

		fmt.Printf("There are %d tags in %s repository.\n", len(tags), repo)

		for tagIndex, tag := range tags {
			tagExists, err := dstRegistry.RepositoryTagExists(repo, tag)
			if err != nil {
				return microerror.Mask(err)
			}
			if tagExists {
				fmt.Printf("\nRepository [%d/%d], tag [%d/%d]: image `%s/%s:%s` already exists.\n", repoIndex+1, len(reposToSync), tagIndex+1, len(tags), r.flag.DstRegistryName, repo, tag)
				continue
			}

			if !tagExists {
				fmt.Printf("\nRepository [%d/%d], tag [%d/%d]: image `%s/%s:%s` is missing.\n", repoIndex+1, len(reposToSync), tagIndex+1, len(tags), r.flag.DstRegistryName, repo, tag)

				err := srcRegistry.PullImage(repo, tag)
				if err != nil {
					return microerror.Mask(err)
				}

				err = registry.RetagImage(repo, tag, sourceRegistryName, r.flag.DstRegistryName)
				if err != nil {
					return microerror.Mask(err)
				}

				err = srcRegistry.RemoveImage(repo, tag)
				if err != nil {
					return microerror.Mask(err)
				}

				err = dstRegistry.PushImage(repo, tag)
				if err != nil {
					return microerror.Mask(err)
				}

				err = dstRegistry.RemoveImage(repo, tag)
				if err != nil {
					return microerror.Mask(err)
				}

			}
		}
	}

	return nil
}
