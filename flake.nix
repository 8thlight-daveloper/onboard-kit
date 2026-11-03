{
  description = "onboard-kit development environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      releaseVersion = nixpkgs.lib.removeSuffix "\n" (builtins.readFile ./VERSION);
      systems = [ "aarch64-darwin" "aarch64-linux" "x86_64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: {
        go = pkgs.go;
        default = pkgs.buildGoModule {
          pname = "onboard-kit";
          version = releaseVersion;
          src = self;
          vendorHash = "sha256-yCNaV9nZP5tW/VMPgg5V5Yqkk6com8stTsJsZh1KaXE=";
          subPackages = [ "binary" ];
          env.CGO_ENABLED = 0;
          doCheck = false;
          ldflags = [ "-X onboard-kit/internal/command/root.version=v${releaseVersion}" ];
          postInstall = ''
            mv $out/bin/binary $out/bin/onboard-kit
            install -Dm644 LICENSE THIRD_PARTY_NOTICES.md -t $out/share/doc/onboard-kit
          '';
          meta = {
            description = "Turns onboarding answers into a plan a newcomer walks through in the terminal";
            license = pkgs.lib.licenses.mit;
            mainProgram = "onboard-kit";
          };
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gopls
            pkgs.just
            pkgs.gh
            pkgs.lychee
            pkgs.markdownlint-cli2
            pkgs.vale
            (pkgs.python3.withPackages (ps: [
              ps.mdformat
              ps.mdformat-gfm
              ps.mdformat-frontmatter
            ]))
          ];
        };
      });
    };
}
