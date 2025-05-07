package go_mod

import (
	"fmt"
	"slices"
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
	visitedModuleFiles map[string]struct{}
}

func NewLanguage() language.Language {
	return &ModuleFiles{
		visitedModuleFiles: make(map[string]struct{}),
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

var IGNORED_FILES = map[string]bool{
	".DS_Store":   true,
	".gitignore":  true,
	"BUILD.bazel": true,
}

// Note: This modifies the input, due to the use of slices.DeleteFunc
// If you don't want the behavior, please clone the input prior to passing it in
func deleteUnwanted(entries []string) []string {
	return slices.DeleteFunc(entries, func(f string) bool {
		_, ignored := IGNORED_FILES[f]
		return strings.HasPrefix(f, ".") || ignored
	})
}

// Add config struct for tracking packages and subdirs

type ModuleFilesConfig struct {
	Packages map[string]bool     // rel path -> is package
	Subdirs  map[string][]string // rel path -> subdirs
}

// Register FinishableLanguage interface
var (
	_ language.Language = &ModuleFiles{}
)

func (mf *ModuleFiles) Configure(c *config.Config, rel string, f *rule.File) {
	fmt.Printf("configure rel: %s\n", rel)
}

// generates the files
func (mf *ModuleFiles) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	fmt.Printf("gen: %s\n", args.Rel)
	fmt.Printf("config.Exts: %v\n", args.Config.Exts["go"])
	// Generate the filegroup for this package if it has the hasRelevantFiles
	var res language.GenerateResult
	srcs := deleteUnwanted(slices.Clone(args.RegularFiles))

	if len(srcs) > 0 {
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
		for rel := range mf.visitedModuleFiles {
			if strings.HasPrefix(rel, args.Rel) {
				// construct the label
				srcs = append(srcs, "//"+rel+":"+TARGET_NAME)

				// Ok we used it, so we can delete it
				delete(mf.visitedModuleFiles, rel)
			}
		}

		r.SetAttr("srcs", srcs)
		r.SetAttr("visibility", []string{"//:__subpackages__"})
		res.Gen = append(res.Gen, r)
		res.Imports = append(res.Imports, []resolve.ImportSpec{})

		// Add this Rel, so it's ancestor can use it.
		// note: this relies on the fact that gazelle is using a depth-first post-order traversal
		mf.visitedModuleFiles[args.Rel] = struct{}{}
	}

	return res
}
