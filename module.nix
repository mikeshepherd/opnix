{ config, lib, pkgs, ... }: let
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

        # Debug: Check if token file exists and is readable
        if [ ! -f ${cfg.tokenFile} ]; then
          echo "Token file ${cfg.tokenFile} does not exist!"
          exit 1
        fi

        if [ ! -r ${cfg.tokenFile} ]; then
          echo "Token file ${cfg.tokenFile} is not readable!"
          exit 1
        fi

        # Debug: Check token content (length and format)
        TOKEN=$(cat ${cfg.tokenFile})
        if [ -z "$TOKEN" ]; then
          echo "Token file is empty!"
          exit 1
        fi

        echo "Token length: ''${#TOKEN} characters"

        # Run the secrets retrieval tool using token file
        ${pkgs.opnix}/bin/opnix \
          -token-file ${cfg.tokenFile} \
          -config ${cfg.configFile} \
          -output ${cfg.outputDir}
      '';
    };
  };
}
