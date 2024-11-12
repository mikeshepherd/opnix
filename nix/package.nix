{ pkgs }:
pkgs.buildGoModule {
  pname = "opnix";
  version = "0.1.0";
  src = ../.;
  vendorHash = "sha256-K8xgmXvhZ4PFUryu9/UsnwZ0Lohi586p1bzqBQBp1jo=";
  subPackages = [ "cmd/opnix" ];
}

