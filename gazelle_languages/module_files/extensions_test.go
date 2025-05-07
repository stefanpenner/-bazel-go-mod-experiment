package module_files

import (
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleFiles_BasicGenerateRules(t *testing.T) {
	ext := NewLanguage().(*ModuleFiles)
	cfg := &config.Config{}

	args := language.GenerateArgs{
		Config:       cfg,
		Rel:          "foo/bar",
		RegularFiles: []string{"file.go", "BUILD.bazel", "go.mod", ".DS_Store"},
	}

	res := ext.GenerateRules(args)

	require.Len(t, res.Gen, 1, "expected 1 rule")
	r := res.Gen[0]
	assert.Equal(t, "filegroup", r.Kind())
	assert.Equal(t, TARGET_NAME, r.Name())

	srcs := r.AttrStrings("srcs")
	wantSrcs := []string{"file.go", "BUILD.bazel", "go.mod"}
	assert.Equal(t, wantSrcs, srcs)

	// we should have visited foo/bar
	assert.Equal(t, []string{"foo/bar"}, ext.visitedModuleFiles.ToSlice())
}

func TestModuleFiles_GenerateRules_Advanced(t *testing.T) {
	ext := NewLanguage().(*ModuleFiles)
	cfg := &config.Config{}

	// example structure:
	// foo/
	//   go.mod
	//   BUILD.bazel
	//   main.go
	//   bar/
	//     alpha/
	//       file.go
	//       BUILD.bazel
	//     beta/
	//       BUILD.bazel
	//       file.go

	// we mimic bazel's post-order traversal invoking GenerateArgs with a subdir first

	// we visit foo/bar/alpha first
	args := language.GenerateArgs{
		Config:       cfg,
		Rel:          "foo/bar/alpha",
		RegularFiles: []string{"alpha.go", "BUILD.bazel"},
	}

	res := ext.GenerateRules(args)

	require.Len(t, res.Gen, 1, "expected 1 rule")
	r := res.Gen[0]
	assert.Equal(t, "filegroup", r.Kind())
	assert.Equal(t, TARGET_NAME, r.Name())

	srcs := r.AttrStrings("srcs")
	wantSrcs := []string{"alpha.go", "BUILD.bazel"}
	assert.Equal(t, wantSrcs, srcs)

	// we should have visited foo/bar/alpha
	assert.Equal(t, []string{"foo/bar/alpha"}, ext.visitedModuleFiles.ToSlice())

	// then foo/bar/beta
	args = language.GenerateArgs{
		Config:       cfg,
		Rel:          "foo/bar/beta",
		RegularFiles: []string{"BUILD.bazel", "beta.go"},
	}
	res = ext.GenerateRules(args)

	require.Len(t, res.Gen, 1, "expected 1 rule")
	r = res.Gen[0]
	assert.Equal(t, "filegroup", r.Kind())
	assert.Equal(t, TARGET_NAME, r.Name())

	srcs = r.AttrStrings("srcs")
	wantSrcs = []string{"BUILD.bazel", "beta.go"}
	assert.Equal(t, wantSrcs, srcs)

	// we should have visited foo/bar/alpha
	assert.Equal(t, []string{"foo/bar/alpha", "foo/bar/beta"}, ext.visitedModuleFiles.ToSlice())

	// then foo
	args = language.GenerateArgs{
		Config:       cfg,
		Rel:          "foo",
		RegularFiles: []string{"BUILD.bazel", "foo.go", "go.mod"},
	}

	res = ext.GenerateRules(args)

	require.Len(t, res.Gen, 1, "expected 1 rule")
	r = res.Gen[0]
	assert.Equal(t, "filegroup", r.Kind())
	assert.Equal(t, TARGET_NAME, r.Name())

	srcs = r.AttrStrings("srcs")
	wantSrcs = []string{"BUILD.bazel", "foo.go", "go.mod", "//foo/bar/alpha:_pkg_", "//foo/bar/beta:_pkg_"}
	assert.Equal(t, wantSrcs, srcs)

	// we should have visited foo/bar/alpha, foo/bar/beta, and foo, but since foo contains foo/bar/{alpha,beta} we expect the contained packages to be removed
	assert.Equal(t, []string{"foo"}, ext.visitedModuleFiles.ToSlice())
}
