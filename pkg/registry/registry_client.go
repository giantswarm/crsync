package registry

type RegistryClient interface {
	Authorize(user, password string) error
	ListRepositories() ([]string, error)
	ListTags(repositry string) ([]string, error)
}
