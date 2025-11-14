package go_mod

import (
	"reflect"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func TestGenerateRulesCollectsLibrariesFromSubdirs(t *testing.T) {
	lang := NewLanguage().(*GoMod)

	childRule := rule.NewRule("go_library", "foo_lib")
	res := lang.GenerateRules(language.GenerateArgs{
		Rel:      "mod/foo",
		OtherGen: []*rule.Rule{childRule},
	})
	if len(res.Gen) != 0 {
		t.Fatalf("expected no rules for subdir without go.mod, got %d", len(res.Gen))
	}

	parentRule := rule.NewRule("go_library", "root_lib")
	res = lang.GenerateRules(language.GenerateArgs{
		Rel:          "mod",
		RegularFiles: []string{"go.mod"},
		OtherGen:     []*rule.Rule{parentRule},
	})

	if len(res.Gen) != 1 {
		t.Fatalf("expected 1 go_mod rule, got %d", len(res.Gen))
	}

	deps := res.Gen[0].AttrStrings("deps")
	want := []string{"//mod/foo:foo_lib", "//mod:root_lib"}
	if !reflect.DeepEqual(deps, want) {
		t.Fatalf("deps mismatch: got %v want %v", deps, want)
	}
}

func TestGenerateRulesSkipsNestedModules(t *testing.T) {
	lang := NewLanguage().(*GoMod)

	childRule := rule.NewRule("go_library", "child_lib")
	res := lang.GenerateRules(language.GenerateArgs{
		Rel:          "mod/submodule",
		RegularFiles: []string{"go.mod"},
		OtherGen:     []*rule.Rule{childRule},
	})

	if len(res.Gen) != 1 {
		t.Fatalf("expected go_mod rule for child module, got %d", len(res.Gen))
	}

	childDeps := res.Gen[0].AttrStrings("deps")
	wantChild := []string{"//mod/submodule:child_lib"}
	if !reflect.DeepEqual(childDeps, wantChild) {
		t.Fatalf("child deps mismatch: got %v want %v", childDeps, wantChild)
	}

	parentRule := rule.NewRule("go_library", "parent_lib")
	res = lang.GenerateRules(language.GenerateArgs{
		Rel:          "mod",
		RegularFiles: []string{"go.mod"},
		OtherGen:     []*rule.Rule{parentRule},
	})

	if len(res.Gen) != 1 {
		t.Fatalf("expected go_mod rule for parent module, got %d", len(res.Gen))
	}

	parentDeps := res.Gen[0].AttrStrings("deps")
	wantParent := []string{"//mod:parent_lib"}
	if !reflect.DeepEqual(parentDeps, wantParent) {
		t.Fatalf("parent deps mismatch: got %v want %v", parentDeps, wantParent)
	}
}
