{
  pkgs ? import <nixpkgs> { },
}: pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls

    libdrm
    linuxHeaders

    # example build dependencies
    pkg-config
    wayland
    libGL
    libgbm
    libdrm
    xorg.libXi
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libX11
    libxkbcommon
  ];
}
