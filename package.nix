{ pkgs }:
pkgs.buildGoModule {
  pname = "opnix";
  version = "0.1.0";
  src = ./.;
  vendorHash = "sha256-owAPNn818xd+KNjSSFiLKeqL1KfDL9Espw6rwCddjbw=";
  subPackages = [ "cmd/opnix" ];
}

