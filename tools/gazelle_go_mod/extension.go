package go_mod

import (
	"flag"
	"fmt"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"

	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// GoMod is the Gazelle extension for go.mod files.
type GoMod struct{}

// NewLanguage creates a new instance of the GoMod extension.
func NewLanguage() language.Language {
	return &GoMod{}
}

// Name returns the name of the language.
func (*GoMod) Name() string {
	return "go_mod"
}

// Kinds returns the kinds of rules this extension generates.
func (*GoMod) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"go_mod": {
			MatchAny:       true,
			NonEmptyAttrs:  map[string]bool{"srcs": true},
			MergeableAttrs: map[string]bool{"srcs": true},
		},
	}
}

// Loads returns the Starlark load statements needed for the rules.
func (*GoMod) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name:    "@//rules:go_mod.bzl",
			Symbols: []string{"go_mod"},
		},
	}
}

// GenerateRules generates rules for go.mod files in a directory.
func (*GoMod) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var srcs []string
	// TODO: this actually needs to work properly:
	// - files in child directories of go.mod should be included if they match the go.mod inclusion pattern
	// - subpackage files need to be included (assuming they match the go.mod inclusion pattern) but will need their BUILD.bazel to have the appropriate rules for visibility setup

	fmt.Printf("Generating rules for go.mod files in %s\n", args.Dir)
	for _, file := range args.RegularFiles {
		fmt.Printf("  - file: %s\n", file)
		if strings.HasSuffix(file, "go.mod") {
			srcs = append(srcs, file)
		}
	}

	var res language.GenerateResult
	if len(srcs) > 0 {
		r := rule.NewRule("go_mod", args.Rel)
		r.SetAttr("go_mod", ":go.mod")
		r.SetAttr("srcs", srcs)
		r.SetAttr("module_path", args.Rel)
		res.Gen = append(res.Gen, r)
		res.Imports = append(res.Imports, nil)
		fmt.Printf("Bingo! %s %v\n", args.Dir, r)
	}
	return res
}

// Configure is a no-op for this simple extension.
func (*GoMod) Configure(c *config.Config, rel string, f *rule.File) {}

// Imports is a no-op since we're not resolving dependencies.
func (*GoMod) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	return nil
}

// Other methods required by the interface (stubs for simplicity).
func (*GoMod) Fix(c *config.Config, f *rule.File) {}
func (*GoMod) KnownDirectives() []string          { return nil }
func (*GoMod) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
}

func (*GoMod) CheckFlags(f *flag.FlagSet, c *config.Config) error {
	return nil
}

func (*GoMod) Embeds(r *rule.Rule, from label.Label) []label.Label { return nil }

func (*GoMod) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {}
