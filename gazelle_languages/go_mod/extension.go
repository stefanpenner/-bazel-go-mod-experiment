package go_mod

import (
	"path"
	"slices"

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
	var res language.GenerateResult

	if !slices.Contains(args.RegularFiles, "go.mod") {
		// no go.mod, no work to be done
		return res
	}

	// _pkg_ is provided by gazelle/languages/module_files
	srcs := []string{":_pkg_"}

	for _, f := range args.Subdirs {
		pkg := path.Join(args.Rel, f)
		srcs = append(srcs, "//"+pkg+":_pkg_")
	}

	r := rule.NewRule("go_mod", "go_mod_zip")

	r.SetAttr("go_mod", ":go.mod")
	r.SetAttr("srcs", srcs)
	r.SetAttr("module_path", args.Rel)

	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, []resolve.ImportSpec{})

	return res
}
