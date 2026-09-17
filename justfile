# Recompute vendorHash in default.nix from go.mod/go.sum without a Nix
# build (scripts/vendor-hash: the module download cache after the
# tree-sitter-baml source replace, NAR-hashed).
vendor-hash:
    bash scripts/vendor-hash

# Bump a first-party dependency to the head of its main branch and
# recompute vendorHash. rev is for bisecting only; landings track main.
bump repo rev="main":
    GOFLAGS= go get github.com/starbaser/{{repo}}@{{rev}}
    GOFLAGS= go mod tidy
    just vendor-hash
