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

      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          librsvg # rsvg-convert: preferred SVG->PNG rasterizer
          jetbrains-mono
        ];
        shellHook = ''
          mkdir -p .direnv/blizzaga-fonts/share/fonts/truetype/blizzaga
          for font in ./font/*.ttf; do
            ln -sf "$PWD/''${font#./}" ".direnv/blizzaga-fonts/share/fonts/truetype/blizzaga/$(basename "$font")"
          done
          export XDG_DATA_DIRS="$PWD/.direnv/blizzaga-fonts/share:${pkgs.jetbrains-mono}/share''${XDG_DATA_DIRS:+:$XDG_DATA_DIRS}"
        '';
      };
    })
    // {
      overlays.default = final: prev: {
        blizzaga = import ./default.nix {pkgs = final;};
      };
    };
}
