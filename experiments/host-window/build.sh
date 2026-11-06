#!/bin/sh
set -eu
cd "$(dirname "$0")/../.."
if ! command -v pkg-config >/dev/null 2>&1 || ! pkg-config --exists sdl2; then
  echo 'SDL2 development files missing. Void Linux: sudo xbps-install -S SDL2-devel pkg-config' >&2
  exit 1
fi
if ! command -v "${CC:-cc}" >/dev/null 2>&1; then
  echo 'C compiler missing: install clang or gcc' >&2
  exit 1
fi
mkdir -p .astelyn/build
# SDL2's pkg-config output is trusted local toolchain metadata.
# shellcheck disable=SC2046
"${CC:-cc}" -std=c11 -O2 -Wall -Wextra \
  $(pkg-config --cflags sdl2) \
  experiments/host-window/window.c -o .astelyn/build/astelyn-host-window \
  $(pkg-config --libs sdl2)
echo 'Built .astelyn/build/astelyn-host-window'
