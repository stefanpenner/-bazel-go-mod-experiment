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
	// Track go_library targets found in subdirectories
	visitedGoLibraries map[string][]string
}

// Register FinishableLanguage interface
var (
	_ language.FinishableLanguage = &GoMod{}
)

func NewLanguage() language.Language {
	return &GoMod{
		visitedGoLibraries: make(map[string][]string),
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
func (gm *GoMod) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var res language.GenerateResult

	// Always record go_library targets from this package for parent packages to use
	// This relies on gazelle's depth-first post-order traversal
	var localLibs []string
	if args.File != nil {
		for _, r := range args.File.Rules {
			if r.Kind() == "go_library" {
				localLibs = append(localLibs, r.Name())
			}
		}
	}
	if len(localLibs) > 0 {
		gm.visitedGoLibraries[args.Rel] = localLibs
	}

	if !slices.Contains(args.RegularFiles, "go.mod") {
		// no go.mod, no work to be done, but we've recorded libraries for parent packages
		return res
	}

	// Find all go_library targets in the current package
	var srcs []string
	if args.File != nil {
		for _, r := range args.File.Rules {
			if r.Kind() == "go_library" {
				name := r.Name()
				srcs = append(srcs, ":"+name)
			}
		}
	}

	// Collect go_library targets from subdirectories that we've visited
	// This relies on gazelle's depth-first post-order traversal
	for rel, libs := range gm.visitedGoLibraries {
		if strings.HasPrefix(rel, args.Rel) && rel != args.Rel {
			// This is a subdirectory, add its go_library targets
			for _, lib := range libs {
				srcs = append(srcs, "//"+rel+":"+lib)
			}
			// Remove it since we've used it
			delete(gm.visitedGoLibraries, rel)
		}
	}

	// If no go_library targets found, return empty result
	if len(srcs) == 0 {
		return res
	}

	r := rule.NewRule("go_mod", "go_mod")

	r.SetAttr("go_mod", ":go.mod")
	r.SetAttr("srcs", srcs)
	r.SetAttr("module_path", args.Rel)

	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, []resolve.ImportSpec{})

	return res
}

// Finish is called after all packages have been processed
func (gm *GoMod) Finish(c *language.Config) {
	// Clean up any remaining visited libraries
	gm.visitedGoLibraries = make(map[string][]string)
}
