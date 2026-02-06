package views

import (
	"strings"
	"testing"
)

func TestRenderStatusBar_StatusOnly(t *testing.T) {
	result := stripANSI(RenderStatusBar(80, "Backup created", "", "n: new | r: refresh"))
	if !strings.Contains(result, "Backup created") {
		t.Error("expected status text")
	}
	if !strings.Contains(result, "n: new | r: refresh") {
		t.Error("expected help text")
	}
}

func TestRenderStatusBar_ErrorOnly(t *testing.T) {
	result := stripANSI(RenderStatusBar(80, "", "Failed", "n: new"))
	if !strings.Contains(result, "Failed") {
		t.Error("expected error text")
	}
	if !strings.Contains(result, "n: new") {
		t.Error("expected help text")
	}
}

func TestRenderStatusBar_ErrorWins(t *testing.T) {
	result := stripANSI(RenderStatusBar(80, "Success", "Error occurred", "help"))
	if !strings.Contains(result, "Error occurred") {
		t.Error("expected error to win")
	}
	if strings.Contains(result, "Success") {
		t.Error("status should not appear when error present")
	}
}

func TestRenderStatusBar_EmptyAll(t *testing.T) {
	result := stripANSI(RenderStatusBar(80, "", "", "help text"))
	if !strings.Contains(result, "help text") {
		t.Error("expected help text")
	}
}

func TestRenderStatusBar_Truncation(t *testing.T) {
	longMsg := strings.Repeat("x", 200)
	result := RenderStatusBar(40, longMsg, "", "help")
	if !strings.Contains(stripANSI(result), "help") {
		t.Error("expected help text")
	}
}
