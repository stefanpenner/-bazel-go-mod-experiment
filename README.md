## Experimenting with Bazel-based Go Module Publishing

A proof-of-concept for 100% Bazel-based Go module publishing. Requiring `go_mod` rules be declared, but those go_mod rules infer the appropriate underlying edges to nested and external packages:

![image](https://github.com/user-attachments/assets/877567a2-6d18-4ba7-86e6-b666c14d6b86)

https://excalidraw.com/#json=f6YI4RFIy_ekfjJ3NBGvI,vZonSZodkQ0jZYRuVAAi4A

### TODO:

- [ ] go_mod also needs to include go_libraries that reside in subpackages, and not directly referenced in the current package.
- [ ] can go_mod rule infer importpath from it's srcs?
- [ ] can go_mod rule infer go.mod location, rather then hardcoding it?
- [ ] version manifest
- [ ] gazelle rule to generate go_mod files
- [ ] relationshipn betweeen publishing rule and go_mod, how does it work, and how do we derive which go_mods have changed, so we know how to version them.

### Algorithm for Module Publishing

1. Use `bazel diff` to find changed targets
2. Use `bazel query` on changed targets to identify `go_mod` rules
3. Identify `go_mod` rules that need publishing
4. Extract module names from the rules
5. Get manifest of current module versions from authoritative source
6. Get new repository version
7. Create updated version manifest by combining current versions with new version
8. Build and publish modules using updated version manifest as volatile input
