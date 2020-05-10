package ownershit

type PermissionsLevel string

const (
	PermissionsAdmin PermissionsLevel = "admin"
	PermissionsRead  PermissionsLevel = "pull"
	PermissionsWrite PermissionsLevel = "push"
)

type Permissions struct {
	Team  string `yaml:"name"`
	ID    int64
	Level PermissionsLevel `yaml:"level"`
}

type PermissionsSettings struct {
	TeamPermissions []*Permissions `yaml:"team"`
	Repositories    []struct {
		Name      string
		Wiki      bool
		Issues    bool
		Projects  bool
		RepoPerms []*Permissions `yaml:"perms"`
	} `yaml:"repositories"`
	Organization string `yaml:"organization"`
}
