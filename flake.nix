{
  description = "1Password secrets integration for NixOS";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
      };
    in {
      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          alejandra
          just
          go
          gopls
          gotools
          go-tools
          golangci-lint
          nil
        ];
      };

      packages.default = import ./nix/package.nix { inherit pkgs; };

      formatter = pkgs.alejandra;
    }) // {
      nixosModules.default = import ./nix/module.nix;

      overlays.default = final: prev: {
        opnix = import ./nix/package.nix { pkgs = final; };
      };
    };
}
