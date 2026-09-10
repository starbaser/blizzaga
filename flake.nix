{
  description = "A tool for generating images of code and terminal output";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    flake-utils.url = "github:numtide/flake-utils";
    tree-sitter-baml = {
    url = "git+file:/home/eigenmage/dev/projects/alloy/crates/tree-sitter-baml?ref=code/fml-grammar&rev=cc27fb58a2a6687d5ee725c19996b4f67e037f08";
      flake = false;
    };
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    tree-sitter-baml,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
      devGoModule = pkgs.runCommand "blizzaga-dev-go-module" {} ''
        mkdir -p "$out"
        cp ${./go.mod} "$out/blizzaga.mod"
        cp ${./go.sum} "$out/blizzaga.sum"
        chmod u+w "$out/blizzaga.mod"
        printf '\nreplace github.com/starbaser/tree-sitter-baml => %s\n' \
          '${tree-sitter-baml}' >> "$out/blizzaga.mod"
      '';
    in {
      packages.default = import ./default.nix {
        inherit pkgs;
        treeSitterBaml = tree-sitter-baml;
      };

      # The committed font/ tree in XDG layout — the same files font.go
      # embeds, so installed and embedded fonts can never diverge. Consumers
      # take this output instead of reaching into the source tree.
      # Provenance: regenerated from iosevka-eigenmage in the dev loop.
      packages.fonts = pkgs.runCommand "blizzaga-fonts" {} ''
        mkdir -p $out/share/fonts/truetype/blizzaga
        cp ${./font}/*.ttf $out/share/fonts/truetype/blizzaga/
      '';

      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          librsvg # rsvg-convert: preferred SVG->PNG rasterizer
          jetbrains-mono
        ];
        shellHook = ''
          export GOFLAGS="-modfile=${devGoModule}/blizzaga.mod''${GOFLAGS:+ $GOFLAGS}"
          export XDG_DATA_DIRS="${self.packages.${system}.fonts}/share:${pkgs.jetbrains-mono}/share''${XDG_DATA_DIRS:+:$XDG_DATA_DIRS}"
        '';
      };
    })
    // {
      overlays.default = final: prev: {
        blizzaga = import ./default.nix {
          pkgs = final;
          treeSitterBaml = tree-sitter-baml;
        };
      };
    };
}
