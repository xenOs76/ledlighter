{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  name = "Go";

  packages = with pkgs; [
    go
    air
    httpie
    goreleaser
    govulncheck
  ];

  CGO_ENABLED = 0;
  COMPOSE_BAKE = "true";

  shellHook = ''
    echo "Exporting GITHUB_TOKEN..."
    export GITHUB_TOKEN=$(cat ~/.config/goreleaser/github_token || echo NOT_FOUND)
  '';
}
