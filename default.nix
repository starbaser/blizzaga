{pkgs}:
let
  fontDataDirs = pkgs.lib.makeSearchPathOutput "out" "share" [
    pkgs.jetbrains-mono
  ];
in
pkgs.buildGoModule {
  name = "blizzaga";
  src = ./.;
  vendorHash = "sha256-okgtlHtTOgNI8PQdExj1pa8HnbQjJVgbTM025u8b/ws=";
  # Tree-sitter's Go bindings include native sources outside their Go package
  # directories, which standard Go vendoring omits.
  proxyVendor = true;
  nativeBuildInputs = [ pkgs.makeWrapper ];

  postInstall = ''
    mkdir -p "$out/share/fonts/truetype/blizzaga"
    cp font/*.ttf "$out/share/fonts/truetype/blizzaga/"
    wrapProgram "$out/bin/blizzaga" \
      --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.librsvg ]} \
      --prefix XDG_DATA_DIRS : "$out/share:${fontDataDirs}"
  '';
}
