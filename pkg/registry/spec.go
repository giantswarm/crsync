package registry

import "context"

const LIST_TAGS_NO_LIMITS = -1

type Interface interface {
	CountTags(ctx context.Context, repo string) (int, error)
	Login(ctx context.Context, user, password string) error
	Logout(ctx context.Context) error
	ListRepositories(ctx context.Context) ([]string, error)
	ListTags(ctx context.Context, repository string, limit int) ([]string, error)
	Name() string
	Pull(ctx context.Context, repo, tag string) error
	Push(ctx context.Context, repo, tag string) error
	RemoveImage(ctx context.Context, repo, tag string) error
}
