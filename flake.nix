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
      in {
        packages.shminer = if target == null then 
          pkgs.writeScriptBin "shminer-unsupported" "echo 'Platform ${system} is not supported by prebuilt binaries.'" 
        else pkgs.stdenv.mkDerivation (finalAttrs: {
          pname = "shminer";
          version = "1.2.0";

          src = pkgs.fetchurl {
            url = target.url;
            hash = target.hash;
          };

          nativeBuildInputs = [ pkgs.makeWrapper ]
            ++ pkgs.lib.optionals pkgs.stdenv.isLinux [ pkgs.autoPatchelfHook pkgs.wrapGAppsHook3 ]
            ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [ pkgs.undmg ];

          buildInputs = pkgs.lib.optionals pkgs.stdenv.isLinux (with pkgs; [
            webkitgtk_4_1
            gtk3
            libGL
            gsettings-desktop-schemas
          ]);

          dontUnpack = !target.isDmg;

          installPhase = if pkgs.stdenv.isLinux then ''
            mkdir -p $out/bin
            cp $src $out/bin/.shminer-raw
            chmod +x $out/bin/.shminer-raw

            makeWrapper $out/bin/.shminer-raw $out/bin/shminer \
              --prefix XDG_DATA_DIRS : "$GSETTINGS_SCHEMAS_PATH" \
              --set WEBKIT_DISABLE_COMPOSITING_MODE "0" \
              --set WEBKIT_FORCE_COMPOSITING_MODE "1" \
              --prefix LD_LIBRARY_PATH : "/usr/lib:/usr/lib32"
          '' else ''
            mkdir -p "$out/Applications"
            cp -r "${target.appName}" "$out/Applications/"
            mkdir -p $out/bin
            ln -s "$out/Applications/${target.appName}/Contents/MacOS/SHMiner" $out/bin/shminer
          '';

          meta = with pkgs.lib; {
            description = "Mining Client for S-UAH cryptocurrency";
            homepage = "https://github.com/OlexiyOdarchuk/Student-Hryvnia-Miner";
            license = licenses.gpl3;
            platforms = builtins.attrNames sysData;
          };
        });

        packages.default = self.packages.${system}.shminer;
      }
    );
}
