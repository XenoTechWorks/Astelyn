#!/bin/sh
set -eu
cd "$(dirname "$0")/../.."
ROOT=${DARLING_ROOT:-/usr/local/libexec/darling}
CLANG=${ASTELYN_CLANG:-clang-21}
LINKER=${ASTELYN_LD64:-ld64.lld}
command -v "$CLANG" >/dev/null || { echo "Missing $CLANG" >&2; exit 1; }
command -v "$LINKER" >/dev/null || { echo "Missing $LINKER" >&2; exit 1; }
test -f "$ROOT/usr/lib/libSystem.B.dylib" || { echo 'Darling libSystem missing' >&2; exit 1; }
mkdir -p .astelyn/build
"$CLANG" -target x86_64-apple-macos11 -nostdinc -c \
  experiments/darwin-bridge/guest_window.c -o .astelyn/build/guest_window.o
"$LINKER" -arch x86_64 -platform_version macos 11.0 11.0 \
  -syslibroot "$ROOT" -L "$ROOT/usr/lib" -lSystem \
  -o .astelyn/build/astelyn-darwin-window-request .astelyn/build/guest_window.o
file .astelyn/build/astelyn-darwin-window-request
