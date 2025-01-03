{
  description = "elvish";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, gomod2nix }:
    let
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        inherit system;
        pkgs = import nixpkgs { inherit system; };
      });
    in
    {
      packages = forAllSystems ({ system, pkgs, ... }:
        let
          buildGoApplication = gomod2nix.legacyPackages.${system}.buildGoApplication;
        in
        rec {
          default = elvish;

          elvish = buildGoApplication {
            name = "elvish";
            src = ./.;
            go = pkgs.go_1_22;
            pwd = ./.;
            subPackages = [ "cmd/elvish" ];
            CGO_ENABLED = 0;
            flags = [
              "-trimpath"
            ];
            ldflags = [
              "-s"
              "-w"
              "-extldflags -static"
            ];
          };
        });

      # `nix develop` provides a shell containing development tools.
      devShell = forAllSystems ({ system, pkgs }:
        pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_22
            gomod2nix.legacyPackages.${system}.gomod2nix
            gopls
          ];
        });

      overlays.default = final: prev: {
        elvish = self.packages.${final.stdenv.system}.elvish;
      };
    };
}

