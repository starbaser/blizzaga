{pkgs}:
pkgs.buildGoModule {
  name = "blizzaga";
  src = ./.;
  vendorHash = "sha256-auggealYd2dgWwga/bwAg0IUYdkOFDCFkMxgIq83juY=";
}
