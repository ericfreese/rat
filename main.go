package main

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"

	rat "github.com/ericfreese/rat/lib"
)

var flags struct {
	cmd  string
	mode string
}

func init() {
	flag.StringVar(&flags.cmd, "cmd", "cat ~/.config/rat/.ratrc", "command to run")
	flag.StringVar(&flags.mode, "mode", "default", "name of mode")

	flag.Parse()
}

func main() {
	var err error

	if err = rat.Init(); err != nil {
		panic(err)
	}

	defer rat.Close()

	var usr *user.User
	usr, err = user.Current()
	if err != nil {
		panic(err)
	}

	configDir := filepath.Join(usr.HomeDir, ".config", "rat")

	rat.SetAnnotatorsDir(filepath.Join(configDir, "annotators"))

	if config, err := os.Open(filepath.Join(configDir, ".ratrc")); err == nil {
		rat.LoadConfig(config)
		config.Close()
	}

	rat.PushPager(rat.NewCmdPager(flags.mode, flags.cmd, rat.Context{}))

	rat.Run()
}
