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

  shellHook = '''';
}
