package quay

type Repository struct {
	IsPublic     bool   `json:"is_public"`
	Name         string `json:"name"`
	LastModified int    `json:"last_modified"`
}

type RepositoriesJSON struct {
	NextPage     string       `json:"next_page"`
	Repositories []Repository `json:"repositories"`
}
