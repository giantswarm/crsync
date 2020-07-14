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

	var srcRegistry *registry.Registry
	{
		registryClientConfig := quay.Config{
			Namespace:    key.Namespace,
			LastModified: r.flag.LastModified,
		}

		registryClient, err := quay.New(registryClientConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		config := registry.Config{
			Name:           sourceRegistryName,
			RegistryClient: registryClient,
		}

		srcRegistry, err = registry.New(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	reposToSync, err := srcRegistry.ListRepositories()
	if err != nil {
		return microerror.Mask(err)
	}

	var registryClient registry.RegistryClient
	{
		switch registryName := r.flag.DstRegistryName; {
		case registryName == "docker.io":
			registryClientConfig := dockerhub.Config{
				Credentials: registry.Credentials{
					User:     r.flag.DstRegistryUser,
					Password: r.flag.DstRegistryPassword,
				},
			}

			registryClient, err = dockerhub.New(registryClientConfig)
			if err != nil {
				return microerror.Mask(err)
			}
		case strings.HasSuffix(registryName, "azurecr.io"):
			registryClientConfig := azurecr.Config{
				Credentials: registry.Credentials{
					User:     r.flag.DstRegistryUser,
					Password: r.flag.DstRegistryPassword,
				},
				RegistryName: r.flag.DstRegistryName,
			}

			registryClient, err = azurecr.New(registryClientConfig)
			if err != nil {
				return microerror.Mask(err)
			}

		default:
			return microerror.Maskf(executionFailedError, "unknown container registry %#q", registryName)
		}
	}

	var dstRegistry *registry.Registry
	{
		config := registry.Config{
			Name: r.flag.DstRegistryName,
			Credentials: registry.Credentials{
				User:     r.flag.DstRegistryUser,
				Password: r.flag.DstRegistryPassword,
			},
			RegistryClient: registryClient,
		}

		dstRegistry, err = registry.New(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	defer dstRegistry.Logout() // nolint
	err = dstRegistry.Login()
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("There are %d repositories to sync.\n", len(reposToSync))

	for repoIndex, repo := range reposToSync {
		srcTags, err := srcRegistry.ListTags(repo)
		if err != nil {
			return microerror.Mask(err)
		}

		dstTags, err := dstRegistry.ListTags(repo)
		if err != nil {
			return microerror.Mask(err)
		}

		tagsToSync := sliceDiff(srcTags, dstTags)

		if len(tagsToSync) == 0 {
			continue
		}

		fmt.Printf("There are %d tags to sync in %s repository.\n", len(tagsToSync), repo)

		for tagIndex, tag := range tagsToSync {

			fmt.Printf("\nRepository [%d/%d], tag [%d/%d]: image `%s/%s:%s`.\n\n", repoIndex+1, len(reposToSync), tagIndex+1, len(tagsToSync), r.flag.DstRegistryName, repo, tag)

			err := srcRegistry.Pull(repo, tag)
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

			err = dstRegistry.Push(repo, tag)
			if err != nil {
				return microerror.Mask(err)
			}

			err = dstRegistry.RemoveImage(repo, tag)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func sliceDiff(s1, s2 []string) []string {
	var result []string

	for _, e1 := range s1 {
		found := false
		for _, e2 := range s2 {
			if e1 == e2 {
				found = true
				break
			}
		}
		if !found {
			result = append(result, e1)
		}
	}

	return result
}
