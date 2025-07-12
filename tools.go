//go:build tools
// +build tools

//go:generate go run github.com/Khan/genqlient
package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/Khan/genqlient"
	_ "go.uber.org/mock/mockgen"
)
