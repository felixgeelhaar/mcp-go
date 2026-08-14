package protocol

import "testing"

func TestNegotiateVersion_InitializeEra(t *testing.T) {
	tests := []struct {
		requested string
		want      string
	}{
		{"", MCPVersion},
		{version20241105, version20241105},
		{MCPVersion, MCPVersion},
		{"2030-01-01", MCPVersion},
		// 2026-07-28 retires initialize: echo the newest initialize-era version.
		{ModernVersion, MCPVersion},
	}
	for _, tt := range tests {
		if got := NegotiateVersion(tt.requested); got != tt.want {
			t.Errorf("NegotiateVersion(%q) = %q, want %q", tt.requested, got, tt.want)
		}
	}
}

func TestSupportedVersions_IncludesModern(t *testing.T) {
	if !IsSupportedVersion(ModernVersion) {
		t.Fatal("SupportedVersions must include 2026-07-28")
	}
	if IsInitializeVersion(ModernVersion) {
		t.Fatal("2026-07-28 must not be an initialize-era version")
	}
	if !IsModernVersion(ModernVersion) {
		t.Fatal("IsModernVersion(ModernVersion) = false")
	}
	alias := DraftVersion //nolint:staticcheck // pin the deprecated alias
	if alias != ModernVersion {
		t.Fatalf("deprecated alias = %q, want %q", alias, ModernVersion)
	}
}

func TestRequiresProtocolVersionHeader(t *testing.T) {
	if RequiresProtocolVersionHeader(version20250326) {
		t.Error("2025-03-26 must not require MCP-Protocol-Version")
	}
	if !RequiresProtocolVersionHeader(version20250618) {
		t.Error("2025-06-18 must require MCP-Protocol-Version")
	}
	if !RequiresProtocolVersionHeader(ModernVersion) {
		t.Error("2026-07-28 must require MCP-Protocol-Version")
	}
}
