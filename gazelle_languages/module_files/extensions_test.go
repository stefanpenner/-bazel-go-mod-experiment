package module_files

import (
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleFiles_GenerateRules(t *testing.T) {
	ext := NewLanguage().(*ModuleFiles)
	cfg := &config.Config{}
	ext.Configure(cfg, "foo/bar", nil)

	args := language.GenerateArgs{
		Config:       cfg,
		Rel:          "foo/bar",
		RegularFiles: []string{"file.go", "BUILD.bazel", "go.mod", "BUILD.bazel", ".DS_Store"},
	}

	res := ext.GenerateRules(args)

	require.Len(t, res.Gen, 1, "expected 1 rule")
	r := res.Gen[0]
	assert.Equal(t, "filegroup", r.Kind())
	assert.Equal(t, TARGET_NAME, r.Name())

	srcs := r.AttrStrings("srcs")
	wantSrcs := []string{"file.go", "BUILD.bazel", "go.mod"}
	assert.Equal(t, wantSrcs, srcs)
}
