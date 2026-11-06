#!/bin/sh
set -eu
ROOT=${DARLING_ROOT:-/usr/local/libexec/darling}
CC=${DARWIN_CC:-clang-21}
LD=${DARWIN_LD:-ld64.lld}
HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
OUT=${ASTELYN_GUI_OUTPUT:-/tmp/astelyn-gui-classes}
OBJ="${OUT}.o"
"$CC" -target x86_64-apple-macos11 -nostdinc -c "$HERE/gui_classes.c" -o "$OBJ"
"$LD" -arch x86_64 -platform_version macos 11.0 11.0 -syslibroot "$ROOT" -L "$ROOT/usr/lib" -L "$ROOT/usr/lib/system" -lSystem -lobjc -o "$OUT" "$OBJ"
file "$OUT"
printf 'Built %s\n' "$OUT"
printf 'To run: darling shell /Volumes/SystemRoot%s\n' "$OUT"
