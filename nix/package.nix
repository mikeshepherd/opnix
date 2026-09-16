{pkgs}:
pkgs.buildGoModule {
  pname = "opnix";
  version = "0.10.1";
  src = ../.;
  vendorHash = "sha256-++BUJ8vDV9N+sqhmKdqopvapDDJiaj/5ZbUfT8mX+cg=";
  subPackages = ["cmd/opnix"];
}
