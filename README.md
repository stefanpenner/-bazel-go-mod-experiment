## Experimenting with Bazel-based Go Module Publishing

A proof-of-concept for 100% Bazel-based Go module publishing.

### Algorithm for Module Publishing

1. Use `bazel diff` to find changed targets
2. Use `bazel query` on changed targets to identify `go_mod` rules
3. Identify `go_mod` rules that need publishing
4. Extract module names from the rules
5. Get manifest of current module versions from authoritative source
6. Get new repository version
7. Create updated version manifest by combining current versions with new version
8. Build and publish modules using updated version manifest as volatile input