package module_files

import (
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// to understand whats going on here please read:
// https://github.com/bazel-contrib/bazel-gazelle/blob/f4f1b2cdee4ac7e452bcf66cadb33429d377965f/language/lang.go#L63

// ModuleFiles is a Gazelle extension for enumerating files that should be packaged.
// This is for bzlmods & go.mods for now, but we may want to decouple in the future.
type ModuleFiles struct {
	language.BaseLang  // see: https://github.com/bazel-contrib/bazel-gazelle/blob/master/language/base.go
	visitedModuleFiles *OrderedSet[string]
}

func NewLanguage() language.Language {
	return &ModuleFiles{
		visitedModuleFiles: NewOrderedSet[string](),
	}
}

const TARGET_NAME = "_pkg_"

func (*ModuleFiles) Name() string {
	return "module_files"
}

// returns the kinds of rules this extension generates.
func (*ModuleFiles) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"filegroup": {
			MatchAny:       true,
			NonEmptyAttrs:  map[string]bool{"srcs": true},
			MergeableAttrs: map[string]bool{"srcs": true},
		},
	}
}

// Register FinishableLanguage interface
var (
	_ language.Language = &ModuleFiles{}
)

func (mf *ModuleFiles) KnownDirectives() []string {
	return []string{"module_files_exclude"}
}

func (mf *ModuleFiles) Configure(c *config.Config, rel string, f *rule.File) {
	if f == nil {
		// somethings this get's called without the rule.File, at which point we can't really proceed so we skip
		// TODO: understand why this happens
		return
	}

	// some sensible defaults
	patterns := []string{".DS_Store", ".bazelignore", ".gitignore", ".bazelrc"}
	if v, ok := c.Exts["module_files_exclude"].([]string); ok {
		patterns = v
	}
	for _, d := range f.Directives {
		if d.Key == "module_files_exclude" {
			patterns = append(patterns, strings.TrimSpace(d.Value))
		}
	}

	c.Exts["module_files_exclude"] = patterns
}

// Helper to filter files for our module_files_exclude directive:
// # gazelle:module_files_exclude patterns
// Otherwise we rely on gazelle's default inclusive/e
func filterModuleFiles(files []string, patterns []string) []string {
	if len(patterns) == 0 {
		return files
	}
	filtered := []string{}
	for _, f := range files {
		ignore := false
		for _, pat := range patterns {
			if matched, _ := filepath.Match(pat, f); matched {
				ignore = true
				break
			}
		}
		if !ignore {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// generates the files
func (mf *ModuleFiles) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var res language.GenerateResult

	patterns, _ := args.Config.Exts["module_files_exclude"].([]string)
	files := filterModuleFiles(args.RegularFiles, patterns)

	srcs := NewOrderedSetFromSlice(files)

	if srcs.Len() > 0 {
		r := rule.NewRule("filegroup", TARGET_NAME)

		// have ancestor BUILD.bazel files track descended ones, even if they are
		// nested several directories deep. This avoids needing a BUILD.bazel at
		// every level
		//
		// note: this relies on gazelles depth-first post-order traversal
		// note: in theory args.Config.Ext should help here, but that approach
		//was more verbose. So unless I fun into a problem, I'll keep it this way.
		//
		// note: I'm not sure a dictionary or array is better here, based on the
		// access pattern it's probably a wash, but I'll keep an eye on things.

		mf.visitedModuleFiles.Range(func(rel string) {
			if strings.HasPrefix(rel, args.Rel) {
				// construct the label
				srcs.Add("//" + rel + ":" + TARGET_NAME)
				// Ok we used it, so we can delete it
				mf.visitedModuleFiles.Remove(rel)
			}
		})

		r.SetAttr("srcs", srcs.ToSlice())
		r.SetAttr("visibility", []string{"//:__subpackages__"})
		res.Gen = append(res.Gen, r)
		res.Imports = append(res.Imports, []resolve.ImportSpec{})

		// Add this Rel, so it's ancestor can use it.
		// note: this relies on the fact that gazelle is using a depth-first post-order traversal
		mf.visitedModuleFiles.Add(args.Rel)
	}

	return res
}
