package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	rat "github.com/ericfreese/rat/lib"
)

var (
	RAT_VERSION = "0.0.2"
)

var flags struct {
	cmd     string
	mode    string
	version bool
}

func init() {
	flag.StringVar(&flags.cmd, "cmd", "cat ~/.config/rat/ratrc", "command to run")
	flag.StringVar(&flags.mode, "mode", "default", "name of mode")
	flag.BoolVar(&flags.version, "version", false, "display version and exit")

	flag.Parse()
}

func main() {
	var err error

	if flags.version {
		fmt.Println(RAT_VERSION)
		return
	}

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
