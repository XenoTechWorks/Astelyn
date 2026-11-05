package gfxconfig

import (
	"reflect"
	"testing"
)

func TestParseVulkanSummary(t *testing.T) {
	input := "GPU0:\n deviceName = AMD Radeon RX 9060 XT\nGPU1:\n deviceName = Intel Graphics\n deviceName = AMD Radeon RX 9060 XT\n"
	expected := []string{"AMD Radeon RX 9060 XT", "Intel Graphics"}
	if got := ParseVulkanSummary(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("got %v", got)
	}
	if got := ParseVulkanSummary("no GPUs"); len(got) != 0 {
		t.Fatalf("unexpected %v", got)
	}
}
