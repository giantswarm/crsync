package sync

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"

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

	var srcRegistry registry.Interface
	{
		c := registry.Config{
			Name:           sourceRegistryName,
			RegistryClient: srcRegistryClient,
		}

		srcRegistry, err = registry.New(c)
		if err != nil {
			return microerror.Mask(err)
		}

		srcRegistry, err = newDecoratedRegistry(srcRegistry)
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

	var dstRegistry registry.Interface
	{
		config := registry.Config{
			Name:           r.flag.DstRegistryName,
			RegistryClient: dstRegistryClient,
		}

		dstRegistry, err = registry.New(config)
		if err != nil {
			return microerror.Mask(err)
		}

		dstRegistry, err = newDecoratedRegistry(dstRegistry)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = r.sync(ctx, srcRegistry, dstRegistry)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) sync(ctx context.Context, srcRegistry, dstRegistry registry.Interface) error {
	var err error

	err = dstRegistry.Login(ctx, r.flag.DstRegistryUser, r.flag.DstRegistryPassword)
	if err != nil {
		return microerror.Mask(err)
	}
	defer func(ctx context.Context) { _ = dstRegistry.Logout(ctx) }(ctx)

	reposToSync, err := srcRegistry.ListRepositories(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("There are %d repositories to sync.\n", len(reposToSync))

	for repoIndex, repo := range reposToSync {
		srcTags, err := srcRegistry.ListTags(ctx, repo)
		if err != nil {
			return microerror.Mask(err)
		}

		dstTags, err := dstRegistry.ListTags(ctx, repo)
		if err != nil {
			return microerror.Mask(err)
		}

		tagsToSync := sliceDiff(srcTags, dstTags)

		if len(tagsToSync) == 0 {
			continue
		}

		fmt.Printf("There are %d tags to sync in %s repository.\n", len(tagsToSync), repo)

		for tagIndex, tag := range tagsToSync {
			job := retagJob{
				ID:   fmt.Sprintf("repository [%d/%d] tag [%d/%d] image = `%s/%s:%s`", repoIndex+1, len(reposToSync), tagIndex+1, len(tagsToSync), r.flag.DstRegistryName, repo, tag),
				Repo: repo,
				Tag:  tag,
			}

			err := r.processRetagJob(ctx, job)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func (r *runner) processRetagJob(ctx context.Context, job retagJob) error {
	err := job.Src.Pull(ctx, job.Repo, job.Tag)
	if err != nil {
		return microerror.Mask(err)
	}

	err = registry.RetagImage(job.Repo, job.Tag, sourceRegistryName, r.flag.DstRegistryName)
	if err != nil {
		return microerror.Mask(err)
	}

	err = job.Src.RemoveImage(ctx, job.Repo, job.Tag)
	if err != nil {
		return microerror.Mask(err)
	}

	err = job.Dst.Push(ctx, job.Repo, job.Tag)
	if err != nil {
		return microerror.Mask(err)
	}

	err = job.Dst.RemoveImage(ctx, job.Repo, job.Tag)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func newDecoratedRegistry(reg registry.Interface) (*registry.DecoratedRegistry, error) {
	c := registry.DecoratedRegistryConfig{
		RateLimiter: registry.DecoratedRegistryConfigRateLimiter{
			ListRepositories: rate.NewLimiter(rate.Every(5*time.Second), 1),
			ListTags:         rate.NewLimiter(rate.Every(1*time.Second), 1),
			Pull:             rate.NewLimiter(rate.Every(100*time.Millisecond), 30),
			Push:             rate.NewLimiter(rate.Every(100*time.Millisecond), 30),
		},
		Underlying: reg,
	}

	r, err := registry.NewDecoratedRegistry(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
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
