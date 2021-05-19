package sync

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	getTagsWorkersNum = 100
	// Docker limits the number of parallel pushes to 5. This limit is
	// 4 because there is no real speed gain when going above that and we
	// put unnecessary pressure on the docker daemon. This gives also one
	// slot left is there is another docker push operation executed on
	// the node.
	retagWorkesNum = 4
	listBurst      = 1
	// Docker limits the number of parallel pushes to 5 anyway.
	pullPushBurst = 10
	// Maximum time between logging out and logging in again.
	loginTTL = 24 * time.Hour
)

type runner struct {
	flag        *flag
	logger      micrologger.Logger
	stdout      io.Writer
	stderr      io.Writer
	lastLoginAt *time.Time

	progressTagsDone   int64
	progressTagsTotal  int64
	progressReposDone  int64
	progressReposTotal int64
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

	// Setup progress printer.
	{
		start := time.Now()
		ticker := time.NewTicker(60 * time.Second)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					fmt.Printf(
						"*** Progress: repositories: [%d/%d] tags: [%d/%d] time elapsed: %s ***\n",
						r.progressReposDone, r.progressReposTotal,
						r.progressTagsDone, r.progressTagsTotal,
						time.Since(start).Round(time.Second),
					)
				}
			}
		}()
		defer ticker.Stop()
	}

	var srcRegistryClient registry.RegistryClient
	{
		switch registryName := r.flag.SrcRegistryName; {
		case registryName == "quay.io":
			c := quay.Config{
				Namespace:                  key.Namespace,
				LastModified:               r.flag.LastModified,
				Token:                      r.flag.QuayAPIToken,
				IncludePrivateRepositories: r.flag.IncludePrivateRepositories,
			}

			srcRegistryClient, err = quay.New(c)
			if err != nil {
				return microerror.Mask(err)
			}

		default:
			return microerror.Maskf(executionFailedError, "unknown container registry %#q", registryName)
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

	if !r.flag.Loop {
		err := r.sync(ctx, srcRegistry, dstRegistry)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}

	if r.flag.MetricsPort != 0 {
		go func() {
			fmt.Printf("Serving metrics at :%d\n", r.flag.MetricsPort)
			http.Handle("/metrics", promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{},
			))
			err := http.ListenAndServe(fmt.Sprintf(":%d", r.flag.MetricsPort), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed serving metrics: %s", microerror.Pretty(microerror.Mask(err), true))
			}
		}()
	} else {
		fmt.Println("Metrics disabled")
	}

	for {
		start := time.Now()

		err := r.sync(ctx, srcRegistry, dstRegistry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nSync error:\n%s\n\n", microerror.Pretty(microerror.Mask(err), true))
			errorsTotal.Inc()
		} else {
			fmt.Printf("\nTook %s\n", time.Since(start))
		}

		time.Sleep(time.Duration(r.flag.SyncInterval) * time.Second)
	}
}

func (r *runner) sync(ctx context.Context, srcRegistry, dstRegistry registry.Interface) error {
	var err error

	fmt.Println()

	if r.lastLoginAt == nil {
		fmt.Printf("Logging in source registry...\n")
		err = srcRegistry.Login(ctx, r.flag.SrcRegistryUser, r.flag.SrcRegistryPassword)
		if err != nil {
			return microerror.Mask(err)
		}

		fmt.Printf("Logging in destination registry...\n")
		err = dstRegistry.Login(ctx, r.flag.DstRegistryUser, r.flag.DstRegistryPassword)
		if err != nil {
			return microerror.Mask(err)
		}
		timeNow := time.Now()
		r.lastLoginAt = &timeNow
	} else {
		fmt.Printf("Already logged in\n")
	}

	defer func(ctx context.Context) {
		fmt.Println()
		if r.lastLoginAt != nil && time.Since(*r.lastLoginAt) >= loginTTL {
			fmt.Printf("Logging out of source registry...\n")
			_ = srcRegistry.Logout(ctx)
			fmt.Printf("Logging out of destination registry...\n")
			_ = dstRegistry.Logout(ctx)
			r.lastLoginAt = nil
		}
	}(ctx)

	// getTagsJobCh channel has buffer 4 times bigger listing tags/repos burst to not starve
	// processing.
	getTagsJobCh := make(chan getTagsJob, listBurst*4)
	getTagsWG := sync.WaitGroup{}

	// retagJobCh channel has buffer 2 times bigger push/pull burst to not starve
	// processing.
	retagJobCh := make(chan retagJob, pullPushBurst*2)
	retagWG := sync.WaitGroup{}

	for i := 0; i < getTagsWorkersNum; i++ {
		getTagsWG.Add(1)
		go func(ctx context.Context) {
			defer getTagsWG.Done()
			r.processGetTagsJobs(ctx, getTagsJobCh, retagJobCh)
		}(ctx)
	}
	for i := 0; i < retagWorkesNum; i++ {
		retagWG.Add(1)
		go func(ctx context.Context) {
			defer retagWG.Done()
			r.processRetagJobs(ctx, retagJobCh)
		}(ctx)
	}

	fmt.Println()
	fmt.Printf("Reading list of repositories to sync from source registry...\n")
	reposToSync, err := srcRegistry.ListRepositories(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	r.progressReposTotal = int64(len(reposToSync))

	fmt.Printf("There are %d repositories to sync.\n", r.progressReposTotal)
	if r.progressReposTotal > 0 {
		fmt.Println()
	}

	for repoIndex, repo := range reposToSync {
		job := getTagsJob{
			Src: srcRegistry,
			Dst: dstRegistry,

			ID:   fmt.Sprintf("Repository [%d/%d] = %#q", repoIndex+1, r.progressReposTotal, repo),
			Repo: repo,
		}

		select {
		case <-ctx.Done():
			return microerror.Mask(ctx.Err())
		case getTagsJobCh <- job:
			// Job added.
		}
	}

	// Wat for getting tags to finish.
	close(getTagsJobCh)
	getTagsWG.Wait()
	// Wat for retagging to finish.
	close(retagJobCh)
	retagWG.Wait()

	return nil
}

func (r *runner) processGetTagsJobs(ctx context.Context, jobCh <-chan getTagsJob, resultCh chan retagJob) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobCh:
			if !ok {
				return
			}

			start := time.Now()

			fmt.Printf("%s: Getting list of tags to sync...\n", job.ID)

			tags, err := r.processGetTagsJob(ctx, job)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: Failed to get list of tags to sync: %s\n", job.ID, microerror.Pretty(microerror.Mask(err), true))
				errorsTotal.Inc()
				continue
			}

			_ = atomic.AddInt64(&r.progressTagsTotal, int64(len(tags)))

			fmt.Printf("%s: Scheduling %d tags to sync...\n", job.ID, len(tags))

			for i, t := range tags {
				j := retagJob{
					Src: job.Src,
					Dst: job.Dst,

					ID:   fmt.Sprintf("%s: Tag [%d/%d] = %#q", job.ID, i+1, len(tags), t),
					Repo: job.Repo,
					Tag:  t,
				}

				select {
				case <-ctx.Done():
					fmt.Fprintf(os.Stderr, "%s: Cancelled while scheduling %d/%d job: %s ***\n", job.ID, i+1, len(tags), microerror.Pretty(microerror.Mask(err), true))
					errorsTotal.Inc()
				case resultCh <- j:
					// ok
				}
			}

			fmt.Printf("%s: Done (took %s)\n", job.ID, time.Since(start).Round(time.Second))
			_ = atomic.AddInt64(&r.progressReposDone, 1)
		}
	}
}

func (r *runner) processRetagJobs(ctx context.Context, jobCh <-chan retagJob) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobCh:
			if !ok {
				return
			}

			start := time.Now()

			fmt.Printf("%s: Retagging...\n", job.ID)

			err := r.processRetagJob(ctx, job)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: Failed to retag: %s\n", job.ID, microerror.Pretty(microerror.Mask(err), true))
				errorsTotal.Inc()
				continue
			}

			fmt.Printf("%s: Done (took %s)\n", job.ID, time.Since(start).Round(time.Second))
			_ = atomic.AddInt64(&r.progressTagsDone, 1)
		}
	}
}

func (r *runner) processGetTagsJob(ctx context.Context, job getTagsJob) ([]string, error) {
	var srcTags, dstTags []string

	eg := new(errgroup.Group)
	eg.Go(func() error {
		var err error
		srcTags, err = job.Src.ListTags(ctx, job.Repo)
		if err == nil {
			tagsTotal.WithLabelValues(job.Src.Name(), job.Repo).Set(float64(len(srcTags)))
		}
		return microerror.Mask(err)
	})
	eg.Go(func() error {
		var err error
		dstTags, err = job.Dst.ListTags(ctx, job.Repo)
		if err == nil {
			tagsTotal.WithLabelValues(job.Dst.Name(), job.Repo).Set(float64(len(dstTags)))
		}
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
		// Try to remove the image by best effort in case of error.
		_ = job.Src.RemoveImage(ctx, job.Repo, job.Tag)
		return microerror.Mask(err)
	}

	err = job.Src.RemoveImage(ctx, job.Repo, job.Tag)
	if err != nil {
		return microerror.Mask(err)
	}

	err = job.Dst.Push(ctx, job.Repo, job.Tag)
	if err != nil {
		// Try to remove the image by best effort in case of error.
		_ = job.Dst.RemoveImage(ctx, job.Repo, job.Tag)
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
			ListRepositories: rate.NewLimiter(rate.Every(5*time.Second), listBurst),
			ListTags:         rate.NewLimiter(rate.Every(1*time.Second), listBurst),
			Pull:             rate.NewLimiter(rate.Every(1*time.Second), pullPushBurst),
			Push:             rate.NewLimiter(rate.Every(1*time.Second), pullPushBurst),
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
