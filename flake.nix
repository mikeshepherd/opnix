{
  description = "A Go project built with Nix";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs = {
    nixpkgs,
    flake-utils,
    gomod2nix,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [gomod2nix.overlays.default];
        };
      in {
        devShells.default = pkgs.mkShell {
          inherit gomod2nix;
          buildInputs = with pkgs; [
            alejandra
            just
            go
            gopls
            gotools
            go-tools
            golangci-lint
            nil
            gomod2nix.packages.${system}.default
          ];
        };

        packages.default = pkgs.buildGoModule {
          pname = "opnix";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-owAPNn818xd+KNjSSFiLKeqL1KfDL9Espw6rwCddjbw=";
          modules = ./gomod2nix.toml;
        };
      }
    );
}
