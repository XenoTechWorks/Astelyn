package gfxconfig

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

type GPUs struct {
	Source  string   `json:"source"`
	Devices []string `json:"devices"`
	Note    string   `json:"note"`
}

// ParseVulkanSummary extracts Vulkan device names without assuming GPU indices
// are stable across reboots or drivers.
func ParseVulkanSummary(output string) []string {
	seen := map[string]bool{}
	names := []string{}
	for _, line := range strings.Split(output, "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || strings.TrimSpace(k) != "deviceName" {
			continue
		}
		name := strings.TrimSpace(v)
		if name != "" && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

func DiscoverGPUs() GPUs {
	r := GPUs{Source: "vulkaninfo --summary", Devices: []string{}, Note: "Inventory only; GPU selection is not implemented for Darling/Roblox"}
	if _, err := exec.LookPath("vulkaninfo"); err != nil {
		r.Note = "vulkaninfo not installed; " + r.Note
		return r
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	b, err := exec.CommandContext(ctx, "vulkaninfo", "--summary").CombinedOutput()
	if err != nil {
		r.Note = "vulkaninfo failed or timed out; " + r.Note
		return r
	}
	r.Devices = ParseVulkanSummary(string(b))
	return r
}
