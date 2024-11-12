# OPNix

OPNix is a NixOS integration tool that enables secure secret management using 1Password during system builds. Inspired by agenix, OPNix allows you to seamlessly incorporate 1Password secrets into your NixOS configuration using service accounts.

## Overview

OPNix bridges the gap between 1Password's secure secret storage and NixOS builds by providing a mechanism to fetch secrets at build time using 1Password service accounts. This allows you to:

- Keep your secrets securely stored in 1Password
- Reference secrets in your NixOS configuration
- Automatically retrieve secrets during system activation
- Maintain security best practices while leveraging NixOS's declarative configuration

## Prerequisites

- NixOS
- A 1Password account with administrative access
- A 1Password service account token
- Basic understanding of NixOS configuration

## Installation

Add OPNix to your NixOS configuration:

```nix
{
  inputs.opnix.url = "github:brizzbuzz/opnix";
  
  outputs = { self, nixpkgs, opnix }: {
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      modules = [
        opnix.nixosModules.default
        ./configuration.nix
      ];
    };
  };
}
```

## Setup

1. Create a 1Password service account and generate a token following the [1Password documentation](https://developer.1password.com/docs/service-accounts/get-started).

2. Store the token securely:
```bash
# Create a secure token file
sudo mkdir -p /run/keys
echo "your-1password-token" | sudo tee /run/keys/op-token >/dev/null
sudo chmod 600 /run/keys/op-token
```

3. Create a secrets configuration file:
```json
{
  "secrets": [
    {
      "path": "mysql/root-password",
      "reference": "op://vault/database/root-password"
    },
    {
      "path": "ssl/private-key",
      "reference": "op://vault/certificates/private-key"
    }
  ]
}
```

4. Enable OPNix in your NixOS configuration:
```nix
{
  services.onepassword-secrets = {
    enable = true;
    tokenFile = "/run/keys/op-token";
    configFile = "/path/to/your/secrets.json";
    outputDir = "/var/lib/opnix/secrets";  # Optional, this is the default
  };
}
```

## Usage

### Secret References

Secrets in 1Password are referenced using the format:
```
op://vault-name/item-name/field-name
```

### Configuration Options

| Option | Type | Description |
|--------|------|-------------|
| `enable` | boolean | Enable/disable the OPNix service |
| `tokenFile` | path | Path to the file containing your 1Password service account token |
| `configFile` | path | Path to your secrets configuration JSON file |
| `outputDir` | string | Directory where secrets will be stored (default: "/var/lib/opnix/secrets") |

### Accessing Secrets

Once configured, secrets will be available in your specified output directory. For example:
```nix
{
  services.mysql = {
    enable = true;
    passwordFile = "/var/lib/opnix/secrets/mysql/root-password";
  };
}
```

## Security Considerations

1. **Token Security**: 
   - Store your token file with appropriate permissions (600)
   - Use `/run/keys` or similar secure locations for token storage
   - Never commit tokens to version control

2. **Service Account Permissions**:
   - Create a dedicated service account with minimal required permissions
   - Regularly rotate service account tokens
   - Monitor service account activity in 1Password audit logs

3. **Secret Access**:
   - The output directory permissions are set to 750 by default
   - Ensure proper file permissions for sensitive secrets
   - Consider using runtime secrets management for highly sensitive data

## Troubleshooting

### Common Issues

1. **Token File Errors**:
   ```
   Token file /run/keys/op-token does not exist!
   ```
   - Ensure the token file exists and has correct permissions
   - Verify the path in your configuration

2. **Authentication Failures**:
   - Verify token validity in 1Password
   - Check service account permissions
   - Ensure token format is correct

3. **Secret Reference Errors**:
   - Verify vault, item, and field names in references
   - Check service account access to referenced vaults
   - Ensure proper JSON format in config file

## Development

### Building from Source

```bash
nix build .#opnix
```

### Running Tests

The project supports both direct nix development shells and direnv for a seamless development experience:

Using nix develop:
```bash
nix develop
go test ./...
```

Using direnv:
```bash
# First time setup
direnv allow

# Tests will then work directly
go test ./...
```

The repository includes a `.envrc` file, so if you have direnv installed and hooked into your shell, you'll automatically enter the development environment when you navigate to the project directory.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT License](LICENSE)

## Credits

- Inspired by [agenix](https://github.com/ryantm/agenix)
- Built with [1Password SDK for Go](https://github.com/1Password/onepassword-sdk-go)

## Related Projects

- [agenix](https://github.com/ryantm/agenix) - Age-encrypted secrets for NixOS
- [sops-nix](https://github.com/Mic92/sops-nix) - Atomic secret provisioning for NixOS
