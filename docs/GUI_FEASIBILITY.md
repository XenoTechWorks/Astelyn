# v0.2.6 — GUI feasibility

`astelyn gui-check --json` checks installed Darling libraries and frameworks. It does **not** attempt to launch a window.

## Runtime class experiment

From the repository root:

```sh
sh experiments/gui-feasibility/build.sh
darling shell /Volumes/SystemRoot/tmp/astelyn-gui-classes
```

The experiment is SDK-free, targets macOS x86_64, and dynamically calls `objc_getClass` and `sel_registerName`.
The marker `ASTELYN_GUI_CLASS_PROBE_OK` means the class probe executed, **not** that AppKit works or a window was displayed. Missing `NSWindow` or `NSApplication` supports our hypothesis that the current Darling environment does not load those classes.

The build script assumes host `/tmp` is mapped as `/Volumes/SystemRoot/tmp` in Darling (verified by the previous Objective-C smoke test). It only creates objects/executables in `/tmp`; no Darling system files are changed.

A Cocoa window requires usable AppKit/Foundation and an event-loop/display bridge. Static file presence alone cannot establish functionality. Do not fabricate the missing system frameworks by symlinking Swift overlays.
