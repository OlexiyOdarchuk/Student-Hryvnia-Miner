{
  description = "SHMiner - Mining Client for S-UAH cryptocurrency";
  
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages.shminer = pkgs.stdenv.mkDerivation {
          pname = "shminer";
          version = "1.1.3";

          src = pkgs.fetchurl {
            url = "https://github.com/OlexiyOdarchuk/Student-Hryvnia-Miner/releases/download/v1.1.3/SHMiner-linux-amd64";
            hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          };

          nativeBuildInputs = [ pkgs.autoPatchelfHook ];

          buildInputs = with pkgs; [
            webkit2gtk
            gtk3
            cairo
            gdk-pixbuf
            glib
            pango
          ];

          unpackPhase = "true";

          installPhase = ''
            mkdir -p $out/bin
            cp $src $out/bin/shminer
            chmod +x $out/bin/shminer
          '';

          meta = with pkgs.lib; {
            description = "Mining Client for S-UAH cryptocurrency";
            homepage = "https://github.com/OlexiyOdarchuk/Student-Hryvnia-Miner";
            license = licenses.gpl3;
            platforms = platforms.linux;
          };
        };

        packages.default = self.packages.${system}.shminer;
      }
    );
}