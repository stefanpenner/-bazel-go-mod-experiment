package go_mod

import (
	"path"
	"slices"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// to understand whats going on here please read:
// https://github.com/bazel-contrib/bazel-gazelle/blob/f4f1b2cdee4ac7e452bcf66cadb33429d377965f/language/lang.go#L63

// GoMod is the Gazelle extension for go.mod files.
type GoMod struct {
	language.BaseLang
	// Track go_library targets by package path
	goLibraries map[string][]string // rel -> []target labels
}

// Register FinishableLanguage interface
var (
	_ language.Language = &GoMod{}
)

func NewLanguage() language.Language {
	return &GoMod{
		goLibraries: make(map[string][]string),
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

// Configure is called before GenerateRules to allow the language to read directives
func (g *GoMod) Configure(c *config.Config, rel string, f *rule.File) {
	// Track go_library targets in this package
	if f != nil {
		var libs []string
		for _, r := range f.Rules {
			if r.Kind() == "go_library" {
				name := r.Name()
				if name != "" {
					libs = append(libs, "//"+rel+":"+name)
				}
			}
		}
		if len(libs) > 0 {
			g.goLibraries[rel] = libs
		}
	}
}

// generates rules for go.mod files in a given bazel package
func (g *GoMod) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var res language.GenerateResult

	// Track go_library targets in the current package
	var currentLibs []string
	if args.File != nil {
		for _, r := range args.File.Rules {
			if r.Kind() == "go_library" {
				name := r.Name()
				if name != "" {
					currentLibs = append(currentLibs, ":"+name)
					// Also track for other packages
					if g.goLibraries[args.Rel] == nil {
						g.goLibraries[args.Rel] = []string{}
					}
					g.goLibraries[args.Rel] = append(g.goLibraries[args.Rel], "//"+args.Rel+":"+name)
				}
			}
		}
	}

	if !slices.Contains(args.RegularFiles, "go.mod") {
		// no go.mod, but we've tracked libraries for potential parent go.mod
		return res
	}

	// Collect all go_library targets in this package and subdirectories
	srcs := currentLibs

	// Find go_library targets in subdirectories (already tracked via Configure)
	for subdir := range g.goLibraries {
		if strings.HasPrefix(subdir, args.Rel+"/") || subdir == args.Rel {
			// This is a subdirectory or the current directory
			if subdir != args.Rel {
				// Add all libraries from this subdirectory
				srcs = append(srcs, g.goLibraries[subdir]...)
			}
		}
	}

	// Also check subdirs that might not have been processed yet
	for _, f := range args.Subdirs {
		pkg := path.Join(args.Rel, f)
		// Check if we've already tracked libraries for this subdir
		if _, exists := g.goLibraries[pkg]; !exists {
			// Use heuristic: directory name = target name
			libName := path.Base(f)
			srcs = append(srcs, "//"+pkg+":"+libName)
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
