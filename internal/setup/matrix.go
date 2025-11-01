package setup

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	osMacOS    = "macos"
	osLinux    = "linux"
	osWindows  = "windows"
	archX86_64 = "x86_64"
	archARM64  = "arm64"

	goOSDarwin  = "darwin"
	goOSLinux   = "linux"
	goOSWindows = "windows"
	goArchAMD64 = "amd64"
	goArchARM64 = "arm64"
)

type MatrixMatcher struct {
	OS   string
	Arch string
}

func NewMatrixMatcher() *MatrixMatcher {
	return &MatrixMatcher{
		OS:   getCurrentOS(),
		Arch: getCurrentArch(),
	}
}

func (m *MatrixMatcher) Matches(config *SetupConfig) (bool, error) {
	if config.Matrix == nil {
		return true, nil
	}

	osMatch := m.matchesOS(config.Matrix.OS)
	archMatch := m.matchesArch(config.Matrix.Arch)

	return osMatch && archMatch, nil
}

func (m *MatrixMatcher) matchesOS(requiredOS []string) bool {
	if len(requiredOS) == 0 {
		return true
	}

	for _, os := range requiredOS {
		if m.normalizeOS(os) == m.OS {
			return true
		}
	}

	return false
}

func (m *MatrixMatcher) matchesArch(requiredArch []string) bool {
	if len(requiredArch) == 0 {
		return true
	}

	for _, arch := range requiredArch {
		if m.normalizeArch(arch) == m.Arch {
			return true
		}
	}

	return false
}

func (m *MatrixMatcher) normalizeOS(os string) string {
	os = strings.ToLower(strings.TrimSpace(os))

	switch os {
	case osMacOS, "mac", goOSDarwin:
		return osMacOS
	case osLinux, "ubuntu", "debian", "centos", "rhel", "fedora":
		return osLinux
	case osWindows, "win":
		return osWindows
	default:
		return os
	}
}

func (m *MatrixMatcher) normalizeArch(arch string) string {
	arch = strings.ToLower(strings.TrimSpace(arch))

	switch arch {
	case archX86_64, goArchAMD64, "x64":
		return archX86_64
	case archARM64, "aarch64", "arm":
		return archARM64
	default:
		return arch
	}
}

func getCurrentOS() string {
	switch runtime.GOOS {
	case goOSDarwin:
		return osMacOS
	case goOSLinux:
		return osLinux
	case goOSWindows:
		return osWindows
	default:
		return runtime.GOOS
	}
}

func getCurrentArch() string {
	switch runtime.GOARCH {
	case goArchAMD64:
		return archX86_64
	case goArchARM64:
		return archARM64
	default:
		return runtime.GOARCH
	}
}

func (m *MatrixMatcher) GetCurrentPlatform() string {
	return fmt.Sprintf("%s/%s", m.OS, m.Arch)
}

func (m *MatrixMatcher) GetSupportedPlatforms(config *SetupConfig) []string {
	if config.Matrix == nil {
		return []string{m.GetCurrentPlatform()}
	}

	var platforms []string

	osList := config.Matrix.OS
	if len(osList) == 0 {
		osList = []string{m.OS}
	}

	archList := config.Matrix.Arch
	if len(archList) == 0 {
		archList = []string{m.Arch}
	}

	for _, os := range osList {
		for _, arch := range archList {
			platforms = append(platforms, fmt.Sprintf("%s/%s",
				m.normalizeOS(os), m.normalizeArch(arch)))
		}
	}

	return platforms
}

func (m *MatrixMatcher) IsCurrentPlatformSupported(config *SetupConfig) (bool, error) {
	matches, err := m.Matches(config)
	if err != nil {
		return false, err
	}
	return matches, nil
}
