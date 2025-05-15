package module_files

import (
	"flag"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/bazel-gazelle/testtools"
	"github.com/bazelbuild/bazel-gazelle/walk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleFiles_BasicGenerateRules(t *testing.T) {
	ext := NewLanguage().(*ModuleFiles)
	cfg := &config.Config{}

	args := language.GenerateArgs{
		Config:       cfg,
		Rel:          "foo/bar",
		RegularFiles: []string{"file.go", "BUILD.bazel", "go.mod"},
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

func TestFilterModuleFiles(t *testing.T) {
	t.Run("no patterns returns all files", func(t *testing.T) {
		files := []string{"a.go", "b.go", "c.txt"}
		patterns := []string{}
		got := filterModuleFiles(files, patterns)
		assert.Equal(t, files, got)
	})

	t.Run("single pattern excludes matching files", func(t *testing.T) {
		files := []string{"a.go", "b.go", "c.txt"}
		patterns := []string{"*.txt"}
		want := []string{"a.go", "b.go"}
		got := filterModuleFiles(files, patterns)
		assert.Equal(t, want, got)
	})

	t.Run("multiple patterns exclude all matches", func(t *testing.T) {
		files := []string{"a.go", "b.go", "c.txt", "d.md"}
		patterns := []string{"*.txt", "*.md"}
		want := []string{"a.go", "b.go"}
		got := filterModuleFiles(files, patterns)
		assert.Equal(t, want, got)
	})

	t.Run("pattern matches no files", func(t *testing.T) {
		files := []string{"a.go", "b.go"}
		patterns := []string{"*.txt"}
		want := []string{"a.go", "b.go"}
		got := filterModuleFiles(files, patterns)
		assert.Equal(t, want, got)
	})

	t.Run("all files excluded", func(t *testing.T) {
		files := []string{"a.go", "b.go"}
		patterns := []string{"*.go"}
		want := []string{}
		got := filterModuleFiles(files, patterns)
		assert.Equal(t, want, got)
	})
}

func TestModuleFiles_EnsureCustomFileExcludeDirectivesWork(t *testing.T) {
	files := []testtools.FileSpec{
		{Path: "BUILD.bazel", Content: `# gazelle:module_files_exclude *.txt`},
		{Path: "foo.go", Content: `package foo`},
		{Path: "bar.txt", Content: `text`},
	}

	dir, cleanup := testtools.CreateFiles(t, files)
	defer cleanup()

	ext := NewLanguage().(*ModuleFiles)

	cfg := config.New()
	cfg.RepoRoot = dir
	cfg.WorkDir = dir

	flagSet := flag.NewFlagSet("test", flag.PanicOnError)

	wext := &walk.Configurer{}

	cexts := []config.Configurer{ext, wext}

	for _, config := range cexts {
		config.RegisterFlags(flagSet, "something", cfg)
		config.CheckFlags(flagSet, cfg)
	}

	dirs := []string{dir}

	var rules []*rule.Rule
	walk.Walk(cfg, cexts, dirs, walk.UpdateDirsMode, func(_dir, rel string, c *config.Config, update bool, f *rule.File, subdirs, regularFiles, genFiles []string) {
		if rel == "" { // root directory
			res := ext.GenerateRules(language.GenerateArgs{
				Config:       c,
				Rel:          rel,
				RegularFiles: regularFiles,
			})
			rules = append(rules, res.Gen...)
		}
	})

	require.Len(t, rules, 1)
	r := rules[0]
	assert.Equal(t, "filegroup", r.Kind())
	srcs := r.AttrStrings("srcs")
	assert.ElementsMatch(t, []string{"BUILD.bazel", "foo.go"}, srcs)
}
