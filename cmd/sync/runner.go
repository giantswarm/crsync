package sync

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/giantswarm/crsync/internal/key"
	"github.com/giantswarm/crsync/pkg/azurecr"
	"github.com/giantswarm/crsync/pkg/dockerhub"
	"github.com/giantswarm/crsync/pkg/quay"
	"github.com/giantswarm/crsync/pkg/registry"
)

const (
	sourceRegistryName = "quay.io"

	pullPushBurst = 30
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

	fmt.Printf("Source registry       = %#q\n", sourceRegistryName)
	fmt.Printf("Destination registry  = %#q\n", r.flag.DstRegistryName)

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

	fmt.Println()
	fmt.Printf("Logging in destination registry...\n")
	err = dstRegistry.Login(ctx, r.flag.DstRegistryUser, r.flag.DstRegistryPassword)
	if err != nil {
		return microerror.Mask(err)
	}
	defer func(ctx context.Context) {
		fmt.Println()
		fmt.Printf("Logging out of destination registry...\n")
		_ = dstRegistry.Logout(ctx)
	}(ctx)

	// Job channel has buffer equal to push/pull burst to not starve
	// processing.
	jobCh := make(chan retagJob, pullPushBurst)

	processingErrCh := make(chan error)
	go func(ctx context.Context) {
		processingErrCh <- r.processRetagJobs(ctx, jobCh)
	}(ctx)
	defer close(processingErrCh)

	fmt.Println()
	fmt.Printf("Reading list of repositories to sync from source registry...\n")
	reposToSync, err := srcRegistry.ListRepositories(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("There are %d repositories to sync.\n", len(reposToSync))
	fmt.Println()

	for repoIndex, repo := range reposToSync {
		job := getTagsJob{
			Src: srcRegistry,
			Dst: dstRegistry,

			ID:   fmt.Sprintf("Repository [%d/%d] = %#q", repoIndex+1, len(reposToSync), repo),
			Repo: repo,
		}

		tagsToSync, err := r.processGetTagsJob(ctx, job)
		if err != nil {
			return microerror.Mask(err)
		}

		for tagIndex, tag := range tagsToSync {
			job := retagJob{
				Src: srcRegistry,
				Dst: dstRegistry,

				ID:   fmt.Sprintf("Repository [%d/%d] = %#q: Tag [%d/%d] = %#q", repoIndex+1, len(reposToSync), repo, tagIndex+1, len(tagsToSync), tag),
				Repo: repo,
				Tag:  tag,
			}

			select {
			case <-ctx.Done():
				return microerror.Mask(ctx.Err())
			case err := <-processingErrCh:
				return microerror.Mask(err)
			case jobCh <- job:
				// Job added.
			}
		}
	}

	close(jobCh)

	// Wait for job processing to finish.
	err = <-processingErrCh
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) processRetagJobs(ctx context.Context, jobCh <-chan retagJob) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return microerror.Mask(ctx.Err())
		case job, ok := <-jobCh:
			if !ok {
				return nil
			}

			wg.Add(1)

			go func(ctx context.Context, job retagJob) {
				defer wg.Done()

				fmt.Printf("%s: Retagging...\n", job.ID)

				err := r.processRetagJob(ctx, job)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: Failed: %s\n", job.ID, microerror.JSON(err))
					return
				}

				fmt.Printf("%s: Retagged\n", job.ID)
			}(ctx, job)
		}
	}
}

func (r *runner) processGetTagsJob(ctx context.Context, job getTagsJob) ([]string, error) {
	var srcTags, dstTags []string

	eg := new(errgroup.Group)
	eg.Go(func() error {
		var err error
		srcTags, err = job.Src.ListTags(ctx, job.Repo)
		return microerror.Mask(err)
	})
	eg.Go(func() error {
		var err error
		dstTags, err = job.Dst.ListTags(ctx, job.Repo)
		return microerror.Mask(err)
	})
	err := eg.Wait()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tags := sliceDiff(srcTags, dstTags)

	return tags, nil
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
			Pull:             rate.NewLimiter(rate.Every(100*time.Millisecond), pullPushBurst),
			Push:             rate.NewLimiter(rate.Every(100*time.Millisecond), pullPushBurst),
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
