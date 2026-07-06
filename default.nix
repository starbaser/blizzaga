{pkgs}:
let
  fontDataDirs = pkgs.lib.makeSearchPathOutput "out" "share" [
    pkgs.jetbrains-mono
  ];
in
pkgs.buildGoModule {
  name = "blizzaga";
  src = ./.;
  vendorHash = "sha256-auggealYd2dgWwga/bwAg0IUYdkOFDCFkMxgIq83juY=";
  nativeBuildInputs = [ pkgs.makeWrapper ];

  postInstall = ''
    mkdir -p "$out/share/fonts/truetype/blizzaga"
    cp font/*.ttf "$out/share/fonts/truetype/blizzaga/"
    wrapProgram "$out/bin/blizzaga" \
      --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.librsvg ]} \
      --prefix XDG_DATA_DIRS : "$out/share:${fontDataDirs}"
  '';
}
