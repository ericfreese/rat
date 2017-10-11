package rat

import (
	"os"
	"os/user"
	"path/filepath"
)

var (
	ConfigDir string
)

func init() {
	xdg_config_path := os.Getenv("XDG_CONFIG_HOME")  // POSIX convention
	if xdg_config_path != "" {
		ConfigDir = filepath.Join(xdg_config_path, "rat")
	} else {
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		ConfigDir = filepath.Join(usr.HomeDir, ".config", "rat")
	}
	SetAnnotatorsDir(filepath.Join(ConfigDir, "annotators"))
}
