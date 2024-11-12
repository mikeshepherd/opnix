{
  description = "1Password secrets integration for NixOS";

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
    self,
    nixpkgs,
    flake-utils,
    gomod2nix,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [gomod2nix.overlays.default];
      };

      src = pkgs.runCommand "opnix-source" {} ''
        mkdir -p $out/cmd/opnix
        mkdir -p $out/internal/{config,onepass,secrets}

        cp ${./cmd/opnix/main.go} $out/cmd/opnix/main.go

        cp ${./internal/config}/*.go $out/internal/config/
        cp ${./internal/onepass}/*.go $out/internal/onepass/
        cp ${./internal/secrets}/*.go $out/internal/secrets/

        cp ${./go.mod} $out/go.mod
        cp ${./go.sum} $out/go.sum
        cp ${./gomod2nix.toml} $out/gomod2nix.toml
      '';
    in {
      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          just
          go
          gotools
          go-tools
          golangci-lint
          gomod2nix.packages.${system}.default
        ];
      };

      packages.default = pkgs.buildGoApplication {
        pname = "opnix";
        version = "0.1.0";
        inherit src;
        modules = ./gomod2nix.toml;
        subPackages = [ "cmd/opnix" ];
      };
    }) // {
      nixosModules.default = import ./module.nix;

      overlays.default = final: prev: {
        opnix = self.packages.${prev.system}.default;
      };
    };
}
