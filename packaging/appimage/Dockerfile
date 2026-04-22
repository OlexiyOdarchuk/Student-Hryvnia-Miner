FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    wget \
    fuse \
    libfuse2 \
    Squashfuse \
    squashfuse-tools \
    fuse2fs \
    libfuse-dev \
    libzstd-dev \
    && rm -rf /var/lib/apt/lists/*

RUN wget -q https://github.com/AppImage/AppImageTool/releases/download/continuous/appimagetool-x86_64.AppImage -O /usr/local/bin/appimagetool && \
    chmod +x /usr/local/bin/appimagetool

RUN mkdir -p /opt/project/AppDir/usr/bin

COPY linux-amd64 /opt/project/AppDir/usr/bin/shminer

RUN chmod +x /opt/project/AppDir/usr/bin/shminer

WORKDIR /opt/project

RUN mkdir -p /opt/project/AppDir/usr/share/icons/hicolor/256x256/apps

RUN cat > /opt/project/AppRun << 'EOF'
#!/bin/bash
set -euo pipefail
APPDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export PATH="${APPDIR}/usr/bin:${PATH}"
export LD_LIBRARY_PATH="${APPDIR}/usr/lib:${APPDIR}/usr/lib/x86_64-linux-gnu:${LD_LIBRARY_PATH:-}"
export XDG_DATA_DIRS="${APPDIR}/usr/share:${XDG_DATA_DIRS:-}"
exec "${APPDIR}/usr/bin/shminer" "${@}"
EOF
RUN chmod +x /opt/project/AppRun

RUN cat > /opt/project/shminer.desktop << 'EOF'
[Desktop Entry]
Name=SHMiner
Exec=shminer
Icon=shminer
Type=Application
Categories=Utility;Finance;
EOF

COPY --from=ubuntu:22.04 /usr/share/icons/hicolor/256x256/apps /opt/project/AppDir/usr/share/icons/hicolor/256x256/apps 2>/dev/null || true

ENV VERSION=unknown

RUN appimagetool AppDir "shminer-${VERSION}.AppImage"