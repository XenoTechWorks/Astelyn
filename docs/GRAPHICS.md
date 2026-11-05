# Graphics manager — Astelyn v0.1.3

Astelyn's Go CLI supports:

```sh
go run ./cmd/astelyn graphics list
go run ./cmd/astelyn graphics get
go run ./cmd/astelyn graphics set auto
go run ./cmd/astelyn graphics set vulkan
go run ./cmd/astelyn graphics set opengl
```

`list` probes the Linux host using `vulkaninfo --summary` and `glxinfo -B`.
`get` and `set` read/write a setting in `$XDG_CONFIG_HOME/astelyn/graphics.json`
(or `~/.config/astelyn/graphics.json`), using an atomic replace.

**IMPORTANT:** This preference is not yet connected to Darling, Metal translation,
or the Roblox Player. It cannot force Roblox to use either API. `list` only
indicates host API diagnostics availability, not macOS runtime compatibility.
