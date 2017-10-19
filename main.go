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
	flag.StringVarP(&flags.cmd, "cmd", "c", "", "command to run (required)")
	flag.StringVarP(&flags.mode, "mode", "m", "default", "name of mode")
	flag.BoolVarP(&flags.version, "version", "v", false, "display version and exit")

	flag.Parse()
}

func validateFlags() bool {
	if len(flags.cmd) == 0 {
		fmt.Fprintln(os.Stderr, "flag 'cmd' is required")
		return false
	}

	return true
}

func main() {
	var err error

	if flags.version {
		fmt.Println(RatVersion)
		return
	}

	if !validateFlags() {
		flag.Usage()
		os.Exit(1)
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
