package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/giantswarm/crsync/internal/key"
	"github.com/giantswarm/crsync/pkg/quayio"
	"github.com/giantswarm/crsync/pkg/registry"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
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

	srcRegistry := registry.Registry{
		Address:    "https://quay.io",
		Name:       "quay.io",
		HttpClient: http.Client{},
	}

	lastModified, err := time.ParseDuration(r.flag.LastModified)
	if err != nil {
		return microerror.Mask(err)
	}
	reposToSync, err := quayio.ListRepositories(key.Namespace, lastModified)
	if err != nil {
		return microerror.Mask(err)
	}

	// docker is specific with urls
	var registryAddress string
	{
		if r.flag.DstRegistryName == "docker.io" {
			registryAddress = "https://index.docker.io"
		} else {
			registryAddress = fmt.Sprintf("https://%s", r.flag.DstRegistryName)
		}
	}
	dstRegistry := registry.Registry{
		Address: registryAddress,
		Name:    r.flag.DstRegistryName,
		Credentials: registry.Credentials{
			User:     r.flag.DstRegistryUser,
			Password: r.flag.DstRegistryPassword,
		},
		HttpClient: http.Client{},
	}

	defer dstRegistry.Logout()
	err = dstRegistry.Login()
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("There are %d repositories to sync.\n", len(reposToSync))

	for _, repo := range reposToSync {
		tags, err := srcRegistry.ListRepositoryTags(repo)
		if err != nil {
			return microerror.Mask(err)
		}

		fmt.Printf("There are %d tags in %s repository.\n", len(tags), repo)

		for i, tag := range tags {
			tagExists, err := dstRegistry.RepositoryTagExists(repo, tag)
			if err != nil {
				return microerror.Mask(err)
			}

			if !tagExists {
				fmt.Printf("\n[%d/%d] Image `%s/%s:%s` is missing.\n", i, len(tags), dstRegistry.Name, repo, tag)

				err := srcRegistry.PullImage(repo, tag)
				if err != nil {
					return microerror.Mask(err)
				}

				err = registry.RetagImage(repo, tag, srcRegistry.Name, dstRegistry.Name)
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

			} else {
				fmt.Printf("\n[%d/%d] Image `%s/%s:%s` already exists.\n", i, len(tags), dstRegistry.Name, repo, tag)
			}
		}
	}

	return nil
}
