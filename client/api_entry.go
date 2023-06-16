package client

type apiCollection[T any] struct {
	Origin  string `json:"origin"`
	Link    string `json:"link"`
	Updated string `json:"updated"`
	Paging  struct {
		Total   int `json:"total"`
		PerPage int `json:"perPage"`
		Offset  int `json:"offset"`
	}
	Entry []ApiEntry[T] `json:"entry"`
}

type ApiEntry[T any] struct {
	Name    string            `json:"name"`
	Id      string            `json:"id"`
	Author  string            `json:"author"`
	ACL     AccessControlList `json:"acl"`
	Content T                 `json:"content"`
}

type AccessControlList struct {
	App     string `json:"app"`
	Owner   string `json:"owner"`
	Sharing string `json:"sharing"`
	Perms   struct {
		Read  []string `json:"read"`
		Write []string `json:"write"`
	}
}
