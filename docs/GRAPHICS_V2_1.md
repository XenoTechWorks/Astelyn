# Graphics Manager v2.1 (Astelyn v0.2.1)

Commands:

- `astelyn graphics gpus` lists discovered Vulkan device names (informational only).
- `astelyn graphics native-run -- vkcube` launches an explicitly chosen **native Linux** test program. The persisted `mangohud` preference enables `MANGOHUD=1` for this program if MangoHud is installed.

Use `astelyn graphics overlay set on` before testing. Exit codes from native test programs are forwarded. GPU selection remains unsupported, and this functionality does not integrate with Darling or macOS Metal. For OpenGL, `mangohud glxgears` is a more direct test.

Do not use this as a sandbox or launch untrusted executables. This tool is for local developer-owned native graphics tests only.
