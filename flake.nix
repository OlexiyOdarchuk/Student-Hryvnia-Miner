{
  description = "SHMiner - Mining Client for S-UAH cryptocurrency";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        linuxDeps = pkgs.lib.optionals pkgs.stdenv.isLinux (with pkgs; [
          webkitgtk_4_1
          gtk3
          libGL
          gsettings-desktop-schemas
          fontconfig
          font-awesome
        ]);

        nativeLinuxDeps = pkgs.lib.optionals pkgs.stdenv.isLinux (with pkgs; [
          pkg-config
          makeWrapper
          autoPatchelfHook
          wrapGAppsHook3
        ]);

        # ==========================================
        # ПАКЕТ 1: Завантаження готового бінарника
        # ==========================================
        sysData = {
          "x86_64-linux" = {
            url = "https://github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/releases/download/v1.2.0/SHMiner-linux-amd64";
            hash = "sha256-440KHKVJgNguFsd9zoz/YdYh1CK9jznZ8pLCcxC7GC8=";
            isDmg = false;
          };
          "x86_64-darwin" = {
            url = "https://github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/releases/download/v1.2.0/SHMiner-darwin-amd64.dmg";
            hash = "sha256-K1yf70AYJrTmwdTUsU4B6pm7nL1QeWubcZQi/ncNDd0=";
            isDmg = true;
            appName = "SHMiner-darwin-amd64.app";
          };
          "aarch64-darwin" = {
            url = "https://github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/releases/download/v1.2.0/SHMiner-darwin-arm64.dmg";
            hash = "sha256-sA7oCkp42uOW9qIfZEcCI+uKt5rLJa2+cG4KjAeELsQ=";
            isDmg = true;
            appName = "SHMiner-darwin-arm64.app";
          };
        };

        target = sysData.${system} or null;

        shminer-bin = if target == null then
          pkgs.writeScriptBin "shminer-unsupported" "echo 'Platform ${system} is not supported by prebuilt binaries.'"
        else pkgs.stdenv.mkDerivation {
          pname = "shminer-bin";
          version = "1.2.0";

          src = pkgs.fetchurl {
            url = target.url;
            hash = target.hash;
          };

          nativeBuildInputs = nativeLinuxDeps ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [ pkgs.undmg ];
          buildInputs = linuxDeps;
          dontUnpack = !target.isDmg;

          installPhase = if pkgs.stdenv.isLinux then ''
            mkdir -p $out/bin
            cp $src $out/bin/.shminer-raw
            chmod +x $out/bin/.shminer-raw

            makeWrapper $out/bin/.shminer-raw $out/bin/shminer \
              --prefix XDG_DATA_DIRS : "$GSETTINGS_SCHEMAS_PATH" \
              --set WEBKIT_DISABLE_COMPOSITING_MODE "0" \
              --set WEBKIT_FORCE_COMPOSITING_MODE "1"
          '' else ''
            mkdir -p "$out/Applications"
            cp -r "${target.appName}" "$out/Applications/"
            mkdir -p $out/bin
            ln -s "$out/Applications/${target.appName}/Contents/MacOS/SHMiner" $out/bin/shminer
          '';
        };

        # ==========================================
        # ПАКЕТ 2: Збірка з сирців (Go + Svelte)
        # ==========================================

        frontend = pkgs.buildNpmPackage {
          pname = "shminer-frontend";
          version = "1.2.0";
          src = ./frontend;

          npmDepsHash = "sha256-Fqvf3jWSAiPBeCy746kzb6FlchysgbbXjAna7RTkSow=";

          buildPhase = "npm run build";
          installPhase = "cp -r dist $out";
        };

        shminer-src = pkgs.buildGoModule {
          pname = "shminer-src";
          version = "1.2.0";
          src = ./.;

          vendorHash = "sha256-GN+4i8I9L+5WD3ayhuybUDjbxqt3YRtf1sFzGEQ1BSg=";

          nativeBuildInputs = nativeLinuxDeps ++ [ pkgs.wails ];
          buildInputs = linuxDeps;
          tags = [ "desktop" "production" "webkit2_41" ];

          preBuild = ''
            rm -rf frontend/dist
            cp -r ${frontend} frontend/dist
          '';

          postInstall = ''
            mv $out/bin/Student-Hryvnia-Miner $out/bin/shminer || true

            wrapProgram $out/bin/shminer \
              --prefix XDG_DATA_DIRS : "$GSETTINGS_SCHEMAS_PATH" \
              --set WEBKIT_DISABLE_COMPOSITING_MODE "0" \
              --set WEBKIT_FORCE_COMPOSITING_MODE "1"
          '';
          meta = {
                      mainProgram = "shminer";
          };
        };

      in {
        packages = {
          bin = shminer-bin; # nix build .#bin
          src = shminer-src; # nix build .#src
          default = shminer-bin; # nix build
        };

        devShells.default = pkgs.mkShell {
          nativeBuildInputs = nativeLinuxDeps;
          buildInputs = linuxDeps ++ [
            pkgs.go
            pkgs.nodejs_22
            pkgs.wails
          ];
        };
      }
    );
}
