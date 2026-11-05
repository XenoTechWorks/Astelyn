# Graphics Manager v2 (Astelyn v0.2.0)

`astelyn graphics overlay status` reports MangoHud discovery and stored preference as JSON.
`astelyn graphics overlay set on` enables the preference; `set off` disables it.
`astelyn graphics performance get` prints the saved settings.

Configuration is stored in `$XDG_CONFIG_HOME/astelyn/performance.json` (fallback `~/.config/astelyn/performance.json`) and is separate from `graphics.json` so existing preferences remain intact.

These are Linux host diagnostics/preferences only. MangoHud may work for native Linux Vulkan/OpenGL software, but Astelyn does **not** yet inject it into Darling, macOS Metal or Roblox. GPU selection is deliberately `auto` until a supported device-selection backend exists. No application is launched or modified by `overlay set`.
