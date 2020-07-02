package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/crsync/internal/key"
	"github.com/giantswarm/crsync/pkg/quayio"
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

	var srcRegistry registry.Registry
	{
		config := registry.Config{
			Name:       sourceRegistryName,
			HttpClient: http.Client{},
		}

		srcRegistry, err = registry.New(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	reposToSync, err := quayio.ListRepositories(key.Namespace, r.flag.LastModified)
	if err != nil {
		return microerror.Mask(err)
	}

	var dstRegistry registry.Registry
	{
		config := registry.Config{
			Name: r.flag.DstRegistryName,
			Credentials: registry.Credentials{
				User:     r.flag.DstRegistryUser,
				Password: r.flag.DstRegistryPassword,
			},
			HttpClient: http.Client{},
		}

		dstRegistry, err = registry.New(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	defer dstRegistry.Logout()
	err = dstRegistry.Login()
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("There are %d repositories to sync.\n", len(reposToSync))

	for repoIndex, repo := range reposToSync {
		tags, err := srcRegistry.ListRepositoryTags(repo)
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
