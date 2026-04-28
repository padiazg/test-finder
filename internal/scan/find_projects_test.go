package scan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FindProjectsFn func(*testing.T, []Project, error)

var (
	checkFindProjects = func(fns ...FindProjectsFn) []FindProjectsFn { return fns }

	checkFindProjectsCount = func(count int) FindProjectsFn {
		return func(t *testing.T, list []Project, _ error) {
			t.Helper()
			assert.Equal(t, count, len(list))
		}
	}
)

func TestFindProjects(t *testing.T) {

	tests := []struct {
		name   string
		opts   *FindProjectsOpts
		checks []FindProjectsFn
	}{
		{
			name:   "single project",
			checks: checkFindProjects(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, e := FindProjects(tt.opts)
			for _, c := range tt.checks {
				c(t, r, e)
			}
		})
	}
}
