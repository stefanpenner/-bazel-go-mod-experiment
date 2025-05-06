package go_mod

import (
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// to understand whats going on here please read:
// https://github.com/bazel-contrib/bazel-gazelle/blob/f4f1b2cdee4ac7e452bcf66cadb33429d377965f/language/lang.go#L63

// GoMod is the Gazelle extension for go.mod files.
type GoMod struct {
	language.BaseLang
	// see: https://github.com/bazel-contrib/bazel-gazelle/blob/master/language/base.go
}

func NewLanguage() language.Language {
	return &GoMod{}
}

func (*GoMod) Name() string {
	return "go_mod"
}

// returns the kinds of rules this extension generates.
func (*GoMod) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"go_mod": {
			MatchAny:       true,
			NonEmptyAttrs:  map[string]bool{"srcs": true},
			MergeableAttrs: map[string]bool{"srcs": true},
		},
	}
}

// returns the Starlark load statements needed for the rules.
func (*GoMod) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name:    "@bazel-go-mod-experiment//rules:go_mod.bzl",
			Symbols: []string{"go_mod"},
		},
	}
}

// generates rules for go.mod files in a given bazel package
func (*GoMod) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var hasGoMod bool
	// TODO: this actually needs to work properly:
	// - files in child directories of go.mod should be included if they match the go.mod inclusion pattern
	// - subpackage files need to be included (assuming they match the go.mod inclusion pattern) but will need their BUILD.bazel to have the appropriate rules for visibility setup

	for _, file := range args.RegularFiles {
		if file == "go.mod" {
			hasGoMod = true
			break
		}
	}

	var res language.GenerateResult

	if hasGoMod {
		r := rule.NewRule("go_mod", "go_mod_zip")
		r.SetAttr("go_mod", ":go.mod")
		r.SetAttr("srcs", []string{})
		r.SetAttr("module_path", args.Rel)
		res.Gen = append(res.Gen, r)
		res.Imports = append(res.Imports, []resolve.ImportSpec{})
	}

	return res
}
