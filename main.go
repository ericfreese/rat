package main

import (
	"fmt"
	"os"
	"path/filepath"

	rat "github.com/ericfreese/rat/lib"
	flag "github.com/spf13/pflag"
)

var (
	RatVersion = "0.0.2"
)

var flags struct {
	cmd     string
	mode    string
	version bool
}

func init() {
	flag.StringVarP(&flags.cmd, "cmd", "c", "", "command to run")
	flag.StringVarP(&flags.mode, "mode", "m", "default", "name of mode")
	flag.BoolVarP(&flags.version, "version", "v", false, "display version and exit")

	flag.Parse()
}

func main() {
	var err error

	if flags.version {
		fmt.Println(RatVersion)
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

	if len(flags.cmd) > 0 {
		rat.PushPager(rat.NewCmdPager(flags.mode, flags.cmd, rat.Context{}))
	} else {
		rat.PushPager(rat.NewReadPager(os.Stdin, "<stdin>", flags.mode, rat.Context{}))
	}

	rat.Run()
}
