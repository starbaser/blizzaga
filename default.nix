{pkgs}:
pkgs.buildGoModule {
  name = "blizzaga";
  src = ./.;
  vendorHash = "sha256-Mn8H0M1F1uulc5drs0FB9NHDiDtVnMVCro/8bCS7dOo=";
}
