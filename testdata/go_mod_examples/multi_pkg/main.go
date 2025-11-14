package multipkg

import (
	"github.com/example/multipkg/alpha"
	"github.com/example/multipkg/beta"
)

func All() []string {
	return []string{alpha.Name(), beta.Name()}
}
