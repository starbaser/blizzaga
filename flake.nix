{
  description = "A tool for generating images of code and terminal output";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      packages.default = import ./default.nix {inherit pkgs;};

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
          export XDG_DATA_DIRS="${self.packages.${system}.fonts}/share:${pkgs.jetbrains-mono}/share''${XDG_DATA_DIRS:+:$XDG_DATA_DIRS}"
        '';
      };
    })
    // {
      overlays.default = final: prev: {
        blizzaga = import ./default.nix {pkgs = final;};
      };
    };
}
