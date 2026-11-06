#!/bin/sh
set -eu
cd "$(dirname "$0")/../.."
if ! command -v pkg-config >/dev/null 2>&1 || ! pkg-config --exists sdl2; then
  echo 'SDL2-devel and pkg-config required (Void: sudo xbps-install -S SDL2-devel pkg-config)' >&2
  exit 1
fi
mkdir -p .astelyn/build
# shellcheck disable=SC2046
"${CC:-cc}" -std=c11 -O2 -Wall -Wextra $(pkg-config --cflags sdl2) \
  experiments/window-bridge/bridge.c -o .astelyn/build/astelyn-window-bridge $(pkg-config --libs sdl2)
printf 'Built .astelyn/build/astelyn-window-bridge\n'
