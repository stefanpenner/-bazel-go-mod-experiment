package go_mod

import (
	"slices"
	"sort"
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
	pending map[string][]string
}

func NewLanguage() language.Language {
	return &GoMod{
		pending: map[string][]string{},
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
func (g *GoMod) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var res language.GenerateResult

	if !slices.Contains(args.RegularFiles, "go.mod") {
		g.recordPending(args)
		return res
	}

	deps := g.collectLibraries(args)
	if len(deps) == 0 {
		return res
	}

	r := rule.NewRule("go_mod", "go_mod_zip")
	r.SetAttr("go_mod", ":go.mod")
	r.SetAttr("module_path", args.Rel)
	r.SetAttr("deps", deps)

	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, []resolve.ImportSpec{})

	return res
}

func (g *GoMod) collectLibraries(args language.GenerateArgs) []string {
	var deps []string

	for _, r := range args.OtherGen {
		if r.Kind() == "go_library" {
			deps = append(deps, labelForRule(args.Rel, r.Name()))
		}
	}

	if pending := g.consumeDescendants(args.Rel); len(pending) > 0 {
		deps = append(deps, pending...)
	}

	deps = dedupeAndSort(deps)
	if len(deps) == 0 {
		return nil
	}
	return deps
}

func (g *GoMod) recordPending(args language.GenerateArgs) {
	local := g.collectLocalOnly(args)
	if len(local) == 0 {
		return
	}
	g.pending[args.Rel] = append(g.pending[args.Rel], local...)
}

func (g *GoMod) collectLocalOnly(args language.GenerateArgs) []string {
	var deps []string
	for _, r := range args.OtherGen {
		if r.Kind() == "go_library" {
			deps = append(deps, labelForRule(args.Rel, r.Name()))
		}
	}
	return deps
}

func (g *GoMod) consumeDescendants(rel string) []string {
	var keys []string
	if rel == "" {
		for key := range g.pending {
			keys = append(keys, key)
		}
	} else {
		prefix := rel + "/"
		for key := range g.pending {
			if strings.HasPrefix(key, prefix) {
				keys = append(keys, key)
			}
		}
	}

	sort.Strings(keys)

	var deps []string
	for _, key := range keys {
		deps = append(deps, g.pending[key]...)
		delete(g.pending, key)
	}
	return deps
}

func labelForRule(rel, name string) string {
	if rel == "" {
		return "//:" + name
	}
	return "//" + rel + ":" + name
}

func dedupeAndSort(values []string) []string {
	if len(values) == 0 {
		return values
	}
	sort.Strings(values)
	result := values[:0]
	var last string
	for i, v := range values {
		if i == 0 || v != last {
			result = append(result, v)
			last = v
		}
	}
	return result
}
