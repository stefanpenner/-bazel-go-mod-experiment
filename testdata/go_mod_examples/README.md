This directory contains standalone Go modules that are used as test data for
the `go_mod` rule. Each module includes a `go_mod` target along with a small
`sh_test` that exercises the staged output to ensure all relevant Go sources
are copied into the loose-files directory.
