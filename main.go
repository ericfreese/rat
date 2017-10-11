package main

import (
	"flag"
	"os"
	"path/filepath"

	rat "github.com/ericfreese/rat/lib"
)

var flags struct {
	cmd  string
	mode string
}

func init() {
	flag.StringVar(&flags.cmd, "cmd", "cat ~/.config/rat/ratrc", "command to run")
	flag.StringVar(&flags.mode, "mode", "default", "name of mode")

	flag.Parse()
}

func main() {
	var err error

	if err = rat.Init(); err != nil {
		panic(err)
	}

	defer rat.Close()

	if config, err := os.Open(filepath.Join(rat.ConfigDir, "ratrc")); err == nil {
		rat.LoadConfig(config)
		config.Close()
	}

	rat.PushPager(rat.NewCmdPager(flags.mode, flags.cmd, rat.Context{}))

	rat.Run()
}
