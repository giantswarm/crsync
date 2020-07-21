package quay

type Repository struct {
	IsPublic     bool   `json:"is_public"`
	Name         string `json:"name"`
	LastModified int    `json:"last_modified"`
}

type RepositoriesJSON struct {
	Repositories []Repository `json:"repositories"`
}
