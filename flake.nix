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
        ];
      };
    })
    // {
      overlays.default = final: prev: {
        blizzaga = import ./default.nix {pkgs = final;};
      };
    };
}
