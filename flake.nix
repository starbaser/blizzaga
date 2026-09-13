{
  description = "A tool for generating images of code and terminal output";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    tree-sitter-baml = {
      url = "github:starbaser/tree-sitter-baml/code/fml-grammar";
      flake = false;
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      tree-sitter-baml,
    }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: nixpkgs.legacyPackages.${system});
      # The one Go binding. default.nix's build and the dev shell share it;
      # advance the toolchain here and nowhere else.
      goFor = forAllSystems (system: nixpkgsFor.${system}.go_1_27);
      # The exact replace default.nix's preBuild applies, so `just vendor-hash`
      # can reproduce the goModules fixed-output derivation without a build.
      sourceReplaces =
        system: "go mod edit -replace=github.com/starbaser/tree-sitter-baml=${tree-sitter-baml}";
      devGoModuleFor = forAllSystems (
        system:
        nixpkgsFor.${system}.runCommand "blizzaga-dev-go-module" { } ''
          mkdir -p "$out"
          cp ${./go.mod} "$out/blizzaga.mod"
          cp ${./go.sum} "$out/blizzaga.sum"
          chmod u+w "$out/blizzaga.mod"
          printf '\nreplace github.com/starbaser/tree-sitter-baml => %s\n' \
            '${tree-sitter-baml}' >> "$out/blizzaga.mod"
        ''
      );
    in
    {
      packages = forAllSystems (system: {
        default = import ./default.nix {
          pkgs = nixpkgsFor.${system};
          go = goFor.${system};
          treeSitterBaml = tree-sitter-baml;
        };

        # The committed font/ tree in XDG layout — the same files font.go
        # embeds, so installed and embedded fonts can never diverge. Consumers
        # take this output instead of reaching into the source tree.
        # Provenance: regenerated from iosevka-eigenmage in the dev loop.
        fonts = nixpkgsFor.${system}.runCommand "blizzaga-fonts" { } ''
          mkdir -p $out/share/fonts/truetype/blizzaga
          cp ${./font}/*.ttf $out/share/fonts/truetype/blizzaga/
        '';
      });

      checks = forAllSystems (system: {
        default = self.packages.${system}.default;
      });

      formatter = forAllSystems (system: nixpkgsFor.${system}.nixfmt);

      lib = forAllSystems (system: {
        sourceReplaces = sourceReplaces system;
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
          go = goFor.${system};
        in
        {
          default = pkgs.mkShell {
            packages = [
              go
              pkgs.gopls
              pkgs.librsvg # rsvg-convert: preferred SVG->PNG rasterizer
              pkgs.jetbrains-mono
            ];
            shellHook = ''
              export GOTOOLCHAIN=local
              export GOFLAGS="-modfile=${devGoModuleFor.${system}}/blizzaga.mod''${GOFLAGS:+ $GOFLAGS}"
              export XDG_DATA_DIRS="${self.packages.${system}.fonts}/share:${pkgs.jetbrains-mono}/share''${XDG_DATA_DIRS:+:$XDG_DATA_DIRS}"
            '';
          };
        }
      );
    }
    // {
      overlays.default = final: prev: {
        blizzaga = import ./default.nix {
          pkgs = final;
          treeSitterBaml = tree-sitter-baml;
        };
      };
    };
}
