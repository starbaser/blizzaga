{
  pkgs,
  treeSitterBaml,
  go ? pkgs.go,
}:
let
  buildGoModule = pkgs.buildGoModule.override { inherit go; };
  fontDataDirs = pkgs.lib.makeSearchPathOutput "out" "share" [
    pkgs.jetbrains-mono
  ];
in
buildGoModule {
  name = "blizzaga";
  src = ./.;
  vendorHash = "sha256-okgtlHtTOgNI8PQdExj1pa8HnbQjJVgbTM025u8b/ws=";
  goSum = ./go.sum;
  # Tree-sitter's Go bindings include native sources outside their Go package
  # directories, which standard Go vendoring omits.
  proxyVendor = true;
  nativeBuildInputs = [ pkgs.makeWrapper ];

  # The Go module records dependency identity while the flake supplies the
  # concrete first-party grammar source pinned in flake.lock.
  preBuild = ''
    go mod edit -replace=github.com/starbaser/tree-sitter-baml=${treeSitterBaml}
  '';

  postInstall = ''
    mkdir -p "$out/share/fonts/truetype/blizzaga"
    cp font/*.ttf "$out/share/fonts/truetype/blizzaga/"
    wrapProgram "$out/bin/blizzaga" \
      --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.librsvg ]} \
      --prefix XDG_DATA_DIRS : "$out/share:${fontDataDirs}"
  '';
}
