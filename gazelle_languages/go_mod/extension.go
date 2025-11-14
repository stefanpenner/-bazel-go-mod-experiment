package go_mod

import (
	"slices"
	"strings"

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
	// Track all go_library targets discovered during traversal
	// Each entry is a full label like "//mod_a/foo:foo"
	allGoLibraries []goLibraryInfo
}

type goLibraryInfo struct {
	label   string // Full label like "//mod_a/foo:foo"
	pkgPath string // Package path like "mod_a/foo"
}

func NewLanguage() language.Language {
	return &GoMod{
		allGoLibraries: []goLibraryInfo{},
	}
}

func (*GoMod) Name() string {
	return "go_mod"
}

// returns the kinds of rules this extension generates.
func (*GoMod) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"go_mod": {
			MatchAny:       true,
			NonEmptyAttrs:  map[string]bool{"deps": true},
			MergeableAttrs: map[string]bool{"deps": true},
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
func (gm *GoMod) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var res language.GenerateResult

	// First, collect go_library targets from the current package
	// This happens for all packages, not just those with go.mod
	if args.File != nil {
		for _, r := range args.File.Rules {
			if r.Kind() == "go_library" {
				label := ":" + r.Name()
				if args.Rel != "" {
					label = "//" + args.Rel + label
				}
				gm.allGoLibraries = append(gm.allGoLibraries, goLibraryInfo{
					label:   label,
					pkgPath: args.Rel,
				})
			}
		}
	}

	// Only generate go_mod rule if there's a go.mod in this package
	if !slices.Contains(args.RegularFiles, "go.mod") {
		return res
	}

	// Collect all go_library targets that belong to this module
	// A go_library belongs to this module if its package path starts with the module path
	deps := []string{}
	modulePrefix := args.Rel
	if modulePrefix != "" {
		modulePrefix = modulePrefix + "/"
	}
	
	for _, lib := range gm.allGoLibraries {
		// Check if this go_library belongs to this module
		if lib.pkgPath == args.Rel || strings.HasPrefix(lib.pkgPath, modulePrefix) {
			deps = append(deps, lib.label)
		}
	}

	r := rule.NewRule("go_mod", "go_mod_zip")

	r.SetAttr("go_mod", ":go.mod")
	r.SetAttr("deps", deps)
	r.SetAttr("module_path", args.Rel)

	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, []resolve.ImportSpec{})

	return res
}
