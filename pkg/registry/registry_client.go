package registry

type RegistryClient interface {
	Authorize(user, password string) error
	CountTags(repository string) (int, error)
	ListRepositories() ([]string, error)
	ListTags(repository string, limit int) ([]string, error)
}
