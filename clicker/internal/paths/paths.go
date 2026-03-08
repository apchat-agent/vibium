package paths

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetCacheDir returns the platform-specific cache directory for Vibium.
// Linux: ~/.cache/vibium/
// macOS: ~/Library/Caches/vibium/
// Windows: %LOCALAPPDATA%\vibium\
func GetCacheDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "linux":
		if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
			baseDir = xdgCache
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(home, ".cache")
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(home, "Library", "Caches")
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			baseDir = localAppData
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(home, "AppData", "Local")
		}
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(home, ".cache")
	}

	return filepath.Join(baseDir, "vibium"), nil
}

// GetChromeForTestingDir returns the directory where Chrome for Testing is cached.
func GetChromeForTestingDir() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "chrome-for-testing"), nil
}

// GetChromeExecutable returns the path to Chrome executable.
// First checks Vibium cache, then falls back to system Chrome/Chromium.
func GetChromeExecutable() (string, error) {
	// First, check CHROME_BIN environment variable
	if chromeBin := os.Getenv("CHROME_BIN"); chromeBin != "" {
		if _, err := os.Stat(chromeBin); err == nil {
			return chromeBin, nil
		}
	}

	// Check Vibium cache first
	cftDir, err := GetChromeForTestingDir()
	if err == nil {
		entries, err := os.ReadDir(cftDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					chromePath := getChromePathInVersion(filepath.Join(cftDir, entry.Name()))
					if _, err := os.Stat(chromePath); err == nil {
						return chromePath, nil
					}
				}
			}
		}
	}

	// Fall back to system Chrome/Chromium
	return findSystemChrome()
}

// GetChromedriverPath returns the path to chromedriver.
// First checks Vibium cache, then falls back to system chromedriver.
func GetChromedriverPath() (string, error) {
	// Check Vibium cache first
	cftDir, err := GetChromeForTestingDir()
	if err == nil {
		entries, err := os.ReadDir(cftDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					driverPath := getChromedriverPathInVersion(filepath.Join(cftDir, entry.Name()))
					if _, err := os.Stat(driverPath); err == nil {
						return driverPath, nil
					}
				}
			}
		}
	}

	// Fall back to system chromedriver
	return findSystemChromedriver()
}

// getChromePathInVersion returns the Chrome executable path within a version directory.
func getChromePathInVersion(versionDir string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(versionDir, "Google Chrome for Testing.app", "Contents", "MacOS", "Google Chrome for Testing")
	case "windows":
		return filepath.Join(versionDir, "chrome.exe")
	default: // linux
		return filepath.Join(versionDir, "chrome")
	}
}

// getChromedriverPathInVersion returns the chromedriver path within a version directory.
func getChromedriverPathInVersion(versionDir string) string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(versionDir, "chromedriver.exe")
	default:
		return filepath.Join(versionDir, "chromedriver")
	}
}

// getPlatformString returns the platform string used by Chrome for Testing.
func getPlatformString() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "mac-arm64"
		}
		return "mac-x64"
	case "windows":
		return "win64"
	default: // linux
		if runtime.GOARCH == "arm64" {
			return "linux-arm64"
		}
		return "linux64"
	}
}

// GetPlatformString is exported for use by the installer.
func GetPlatformString() string {
	return getPlatformString()
}

// GetDaemonDir returns the directory for daemon files (socket, PID).
// This reuses the cache directory since daemon files are ephemeral.
func GetDaemonDir() (string, error) {
	return GetCacheDir()
}

// GetSocketPath returns the platform-specific socket path for the daemon.
// macOS/Linux: ~/Library/Caches/vibium/vibium.sock or ~/.cache/vibium/vibium.sock
// Windows: \\.\pipe\vibium (named pipe)
func GetSocketPath() (string, error) {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\vibium`, nil
	}
	dir, err := GetDaemonDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vibium.sock"), nil
}

// GetPIDPath returns the path to the daemon PID file.
func GetPIDPath() (string, error) {
	dir, err := GetDaemonDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vibium.pid"), nil
}

// findSystemChrome searches for system Chrome/Chromium installations.
func findSystemChrome() (string, error) {
	// Common Chrome/Chromium executable names by platform
	var candidates []string

	switch runtime.GOOS {
	case "windows":
		candidates = []string{
			"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
		}
	case "darwin":
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	default: // linux
		candidates = []string{
			"google-chrome",
			"google-chrome-stable",
			"google-chrome-unstable",
			"chromium",
			"chromium-browser",
			"chromium",
		}
	}

	// Check candidates by path first
	for _, candidate := range candidates {
		// Check if it's an absolute path
		isAbs := filepath.IsAbs(candidate)
		if !isAbs {
			// It's a command name, use exec.LookPath
			if path, err := exec.LookPath(candidate); err == nil {
				return path, nil
			}
		} else {
			// It's an absolute path
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}

	return "", os.ErrNotExist
}

// findSystemChromedriver searches for system chromedriver installations.
func findSystemChromedriver() (string, error) {
	// Common chromedriver names
	candidates := []string{
		"chromedriver",
		"chromedriver-linux64",
		"chromedriver-mac-arm64",
		"chromedriver-mac-x64",
		"chromedriver.exe",
	}

	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}

	return "", os.ErrNotExist
}

// GetScreenshotDir returns the platform-specific default directory for screenshots.
// macOS: ~/Pictures/Vibium/
// Linux: ~/Pictures/Vibium/
// Windows: %USERPROFILE%\Pictures\Vibium\
func GetScreenshotDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		// Windows uses Pictures folder in user profile
		return filepath.Join(home, "Pictures", "Vibium"), nil
	default:
		// macOS and Linux use ~/Pictures/Vibium
		return filepath.Join(home, "Pictures", "Vibium"), nil
	}
}
