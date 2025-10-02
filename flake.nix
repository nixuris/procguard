{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs =
    {
      self,
      nixpkgs,
    }:
    let
      system = "x86_64-linux"; # or "aarch64-darwin" for M1/M2 Macs
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          golangci-lint
          cmake
          nodejs_24
        ];
        shellHook = ''
          go env -w GOPATH=$HOME/.local/share/go
          export NPM_CONFIG_PREFIX="$HOME/.local"
          export PATH="$HOME/.local/bin:$PATH"
        '';
      };
    };
}
