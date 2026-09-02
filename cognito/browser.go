package cognito

import (
	"os/exec"
	"runtime"
)

// browserCommand returns the command and arguments that open rawURL in the
// platform's default browser.
func browserCommand(goos, rawURL string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{rawURL}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", rawURL}
	default:
		return "xdg-open", []string{rawURL}
	}
}

// openBrowser launches the platform's default browser. This is the one
// genuinely untestable function in the package - a unit test must not spawn a
// real browser - so it is kept to a single statement over browserCommand,
// which is tested. Callers that need a seam set Config.OpenBrowser instead.
func openBrowser(rawURL string) error {
	name, args := browserCommand(runtime.GOOS, rawURL)
	return exec.Command(name, args...).Start()
}
