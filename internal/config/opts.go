package config

import (
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"
)

type opts struct {
	Args struct {
		Prog       string
		BrokerName string
	} `positional-args:"yes" required:"yes"`
	ParallelUpgrades int `short:"p" long:"parallel" default:"10" description:"number of upgrades to perform in parallel"`
}

func Parse(args []string) (string, int, error) {
	var o opts
	extra, err := newParser(&o).ParseArgs(args)
	switch {
	case err != nil:
		return "", 0, err
	case len(extra) != 0:
		return "", 0, fmt.Errorf("extraneous arguments: %s\n\n%s", strings.Join(extra, " "), Help())
	default:
		return o.Args.BrokerName, o.ParallelUpgrades, nil
	}
}

func Help() string {
	var b strings.Builder
	newParser(&opts{}).WriteHelp(&b)
	return b.String()
}

func newParser(o *opts) *flags.Parser {
	return flags.NewParser(&o, flags.HelpFlag|flags.PrintErrors)
}
