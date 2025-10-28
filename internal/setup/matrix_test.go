package setup

import (
	"testing"
)

func TestMatrixMatcher_Matches_NoMatrix(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
		Name:    "No Matrix",
	}

	matches, err := matcher.Matches(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !matches {
		t.Error("Expected config without matrix to always match")
	}
}

func TestMatrixMatcher_Matches_CurrentOS(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			OS: []string{matcher.OS},
		},
	}

	matches, err := matcher.Matches(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !matches {
		t.Errorf("Expected config to match current OS: %s", matcher.OS)
	}
}

func TestMatrixMatcher_Matches_CurrentArch(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			Arch: []string{matcher.Arch},
		},
	}

	matches, err := matcher.Matches(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !matches {
		t.Errorf("Expected config to match current arch: %s", matcher.Arch)
	}
}

func TestMatrixMatcher_Matches_BothOSAndArch(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			OS:   []string{matcher.OS},
			Arch: []string{matcher.Arch},
		},
	}

	matches, err := matcher.Matches(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !matches {
		t.Errorf("Expected config to match current platform: %s/%s", matcher.OS, matcher.Arch)
	}
}

func TestMatrixMatcher_DoesNotMatch_DifferentOS(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			OS: []string{"different-os"},
		},
	}

	matches, err := matcher.Matches(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if matches {
		t.Error("Expected config with different OS to not match")
	}
}

func TestMatrixMatcher_DoesNotMatch_DifferentArch(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			Arch: []string{"different-arch"},
		},
	}

	matches, err := matcher.Matches(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if matches {
		t.Error("Expected config with different arch to not match")
	}
}

func TestMatrixMatcher_OSNormalization(t *testing.T) {
	matcher := NewMatrixMatcher()

	testCases := []struct {
		input    string
		expected string
	}{
		{"macos", "macos"},
		{"mac", "macos"},
		{"darwin", "macos"},
		{"MACOS", "macos"},
		{"linux", "linux"},
		{"ubuntu", "linux"},
		{"debian", "linux"},
		{"centos", "linux"},
		{"rhel", "linux"},
		{"fedora", "linux"},
		{"LINUX", "linux"},
		{"windows", "windows"},
		{"win", "windows"},
		{"WINDOWS", "windows"},
		{"unknown", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := matcher.normalizeOS(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestMatrixMatcher_ArchNormalization(t *testing.T) {
	matcher := NewMatrixMatcher()

	testCases := []struct {
		input    string
		expected string
	}{
		{"x86_64", "x86_64"},
		{"amd64", "x86_64"},
		{"x64", "x86_64"},
		{"X86_64", "x86_64"},
		{"arm64", "arm64"},
		{"aarch64", "arm64"},
		{"arm", "arm64"},
		{"ARM64", "arm64"},
		{"unknown", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := matcher.normalizeArch(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestMatrixMatcher_MatchesWithNormalization(t *testing.T) {
	matcher := NewMatrixMatcher()

	testCases := []struct {
		name     string
		osList   []string
		archList []string
		expected bool
	}{
		{
			name:     "macos variants",
			osList:   []string{"mac", "darwin", "macos"},
			archList: []string{matcher.Arch},
			expected: matcher.OS == "macos",
		},
		{
			name:     "linux variants",
			osList:   []string{"ubuntu", "debian", "linux"},
			archList: []string{matcher.Arch},
			expected: matcher.OS == "linux",
		},
		{
			name:     "windows variants",
			osList:   []string{"win", "windows"},
			archList: []string{matcher.Arch},
			expected: matcher.OS == "windows",
		},
		{
			name:     "x86_64 variants",
			osList:   []string{matcher.OS},
			archList: []string{"amd64", "x64", "x86_64"},
			expected: matcher.Arch == "x86_64",
		},
		{
			name:     "arm64 variants",
			osList:   []string{matcher.OS},
			archList: []string{"aarch64", "arm", "arm64"},
			expected: matcher.Arch == "arm64",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &SetupConfig{
				Version: 1,
				Matrix: &Matrix{
					OS:   tc.osList,
					Arch: tc.archList,
				},
			}

			matches, err := matcher.Matches(config)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if matches != tc.expected {
				t.Errorf("Expected %v, got %v for OS: %v, Arch: %v",
					tc.expected, matches, tc.osList, tc.archList)
			}
		})
	}
}

func TestMatrixMatcher_GetCurrentPlatform(t *testing.T) {
	matcher := NewMatrixMatcher()
	platform := matcher.GetCurrentPlatform()

	expected := matcher.OS + "/" + matcher.Arch
	if platform != expected {
		t.Errorf("Expected %s, got %s", expected, platform)
	}
}

func TestMatrixMatcher_GetSupportedPlatforms_NoMatrix(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
	}

	platforms := matcher.GetSupportedPlatforms(config)
	expected := []string{matcher.GetCurrentPlatform()}

	if len(platforms) != len(expected) {
		t.Errorf("Expected %d platforms, got %d", len(expected), len(platforms))
	}
	if platforms[0] != expected[0] {
		t.Errorf("Expected %s, got %s", expected[0], platforms[0])
	}
}

func TestMatrixMatcher_GetSupportedPlatforms_WithMatrix(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			OS:   []string{"macos", "linux"},
			Arch: []string{"x86_64", "arm64"},
		},
	}

	platforms := matcher.GetSupportedPlatforms(config)
	expected := []string{
		"macos/x86_64",
		"macos/arm64",
		"linux/x86_64",
		"linux/arm64",
	}

	if len(platforms) != len(expected) {
		t.Errorf("Expected %d platforms, got %d", len(expected), len(platforms))
	}

	for _, expectedPlatform := range expected {
		found := false
		for _, platform := range platforms {
			if platform == expectedPlatform {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected platform %s not found in %v", expectedPlatform, platforms)
		}
	}
}

func TestMatrixMatcher_GetSupportedPlatforms_PartialMatrix(t *testing.T) {
	matcher := NewMatrixMatcher()
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			OS: []string{"macos", "linux"},
		},
	}

	platforms := matcher.GetSupportedPlatforms(config)
	expected := []string{
		"macos/" + matcher.Arch,
		"linux/" + matcher.Arch,
	}

	if len(platforms) != len(expected) {
		t.Errorf("Expected %d platforms, got %d", len(expected), len(platforms))
	}

	for _, expectedPlatform := range expected {
		found := false
		for _, platform := range platforms {
			if platform == expectedPlatform {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected platform %s not found in %v", expectedPlatform, platforms)
		}
	}
}

func TestMatrixMatcher_IsCurrentPlatformSupported(t *testing.T) {
	matcher := NewMatrixMatcher()

	testCases := []struct {
		name     string
		config   *SetupConfig
		expected bool
	}{
		{
			name: "no matrix",
			config: &SetupConfig{
				Version: 1,
			},
			expected: true,
		},
		{
			name: "matches current platform",
			config: &SetupConfig{
				Version: 1,
				Matrix: &Matrix{
					OS:   []string{matcher.OS},
					Arch: []string{matcher.Arch},
				},
			},
			expected: true,
		},
		{
			name: "does not match current platform",
			config: &SetupConfig{
				Version: 1,
				Matrix: &Matrix{
					OS:   []string{"different-os"},
					Arch: []string{"different-arch"},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			supported, err := matcher.IsCurrentPlatformSupported(tc.config)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if supported != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, supported)
			}
		})
	}
}
