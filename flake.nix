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
        lib = pkgs.lib;

        # Єдине джерело версії — wails.json, той самий, що читає релізний CI.
        version = (builtins.fromJSON (builtins.readFile ./wails.json)).version;

        linuxDeps = lib.optionals pkgs.stdenv.isLinux (with pkgs; [
          webkitgtk_4_1
          gtk3
          libGL
          gsettings-desktop-schemas
          fontconfig
          font-awesome
        ]);

        nativeLinuxDeps = lib.optionals pkgs.stdenv.isLinux (with pkgs; [
          pkg-config
          wrapGAppsHook3
        ]);

        # Ядро перебору на Rust. Залежності фіксує Cargo.lock,
        # тож окремого хешу, який треба оновлювати руками, тут немає.
        minerCore = pkgs.rustPlatform.buildRustPackage {
          pname = "shminer-core";
          inherit version;
          src = ./backend/internal/miner/rust_miner_core;
          cargoLock.lockFile = ./backend/internal/miner/rust_miner_core/Cargo.lock;
        };

        frontend = pkgs.buildNpmPackage {
          pname = "shminer-frontend";
          inherit version;
          src = ./frontend;

          # Змінюється лише разом із frontend/package-lock.json.
          npmDepsHash = "sha256-Fqvf3jWSAiPBeCy746kzb6FlchysgbbXjAna7RTkSow=";

          buildPhase = "npm run build";
          installPhase = "cp -r dist $out";
        };

        shminer = pkgs.buildGoModule {
          pname = "shminer";
          inherit version;
          src = ./.;

          # Змінюється лише разом із go.sum.
          vendorHash = "sha256-q5KxUULP4OtJWJ8Ed9Ix3Fwhhi5QV1mHmfAc8RCXoRo=";

          nativeBuildInputs = nativeLinuxDeps;
          buildInputs = linuxDeps;
          tags = [ "desktop" "production" "rustcore" ]
            ++ lib.optionals pkgs.stdenv.isLinux [ "webkit2_41" ];
          ldflags = [ "-s" "-w" ];

          # Статична бібліотека ядра лягає туди, де її шукає cgo
          # (див. searcher_rust.go та scripts/build-rust-core.sh).
          preBuild = ''
            rm -rf frontend/dist
            cp -r ${frontend} frontend/dist

            coreLib=backend/internal/miner/rust_miner_core/lib/${pkgs.go.GOOS}_${pkgs.go.GOARCH}
            mkdir -p "$coreLib"
            cp ${minerCore}/lib/librust_miner_core.a "$coreLib/"
          '';

          postInstall = ''
            mv $out/bin/Student-Hryvnia-Miner $out/bin/shminer
          '';

          # Обгортку робить wrapGAppsHook3, тут лише додаємо до неї змінні.
          preFixup = lib.optionalString pkgs.stdenv.isLinux ''
            gappsWrapperArgs+=(
              --set WEBKIT_DISABLE_COMPOSITING_MODE "0"
              --set WEBKIT_FORCE_COMPOSITING_MODE "1"
            )
          '';

          meta = {
            description = "Mining Client for S-UAH cryptocurrency";
            mainProgram = "shminer";
            license = lib.licenses.gpl3Only;
          };
        };

      in {
        packages = {
          default = shminer; # nix build
          src = shminer;     # стара назва, лишається для сумісності
          core = minerCore;  # nix build .#core
        };

        devShells.default = pkgs.mkShell {
          nativeBuildInputs = nativeLinuxDeps;
          buildInputs = linuxDeps ++ [
            pkgs.go
            pkgs.nodejs_22
            pkgs.wails
            pkgs.cargo
            pkgs.rustc
          ];
        };
      }
    );
}
