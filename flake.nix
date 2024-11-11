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
  }: let
    nixosModule = {config, lib, pkgs, ...}: let
      cfg = config.services.onepassword-secrets;
    in {
      options.services.onepassword-secrets = {
        enable = lib.mkEnableOption "1Password secrets integration";

        tokenFile = lib.mkOption {
          type = lib.types.path;
          description = ''
            Path to file containing the 1Password service account token.
            The file should contain only the token and should have appropriate permissions (600).

            Example:
              Create token file: echo "your-token" > /run/keys/op-token
              Set permissions: chmod 600 /run/keys/op-token
          '';
        };

        configFile = lib.mkOption {
          type = lib.types.path;
          description = "Path to secrets configuration file";
        };

        outputDir = lib.mkOption {
          type = lib.types.str;
          default = "/run/secrets";
          description = "Directory to store retrieved secrets";
        };
      };

      config = lib.mkIf cfg.enable {
        system.activationScripts.onepassword-secrets = {
          deps = [];
          text = ''
            # Ensure output directory exists with correct permissions
            mkdir -p ${cfg.outputDir}
            chmod 750 ${cfg.outputDir}

            # Run the secrets retrieval tool using token file
            ${self.packages.${pkgs.system}.default}/bin/opnix \
              -token-file ${cfg.tokenFile} \
              -config ${cfg.configFile} \
              -output ${cfg.outputDir}
          '';
        };
      };
    };
  in
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [gomod2nix.overlays.default];
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

        # Add formatter
        formatter = pkgs.alejandra;
      }
    ) // {
      # Platform-independent outputs
      nixosModules.default = nixosModule;
      nixosModule = nixosModule; # For compatibility with traditional imports

      # Provide an overlay for the package
      overlays.default = final: prev: {
        opnix = self.packages.${prev.system}.default;
      };

      # Add checks that run on CI
      checks = builtins.mapAttrs (system: pkgs: {
        # You can add more checks here
        default = self.packages.${system}.default;
      }) nixpkgs.legacyPackages;
    };
}
