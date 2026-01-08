package secure

import (
	"fmt"
	"os/exec"
)

// Patch updates the given module to a secure version and tidies the module file.
func Patch(module, secureVersion string) error {
	if err := exec.Command("go", "get", fmt.Sprintf("%s@%s", module, secureVersion)).Run(); err != nil {
		return err
	}
	return exec.Command("go", "mod", "tidy").Run()
}
