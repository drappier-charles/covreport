// Package reporter provides functions for generating coverage reports.
package reporter

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/drappier-charles/covreport/reporter/config"
	"github.com/drappier-charles/covreport/reporter/internal"
)

// Report generates a coverage report using the given configuration.
func Report(cfg *config.Config) error {
	gp := internal.NewGoProject(cfg.Root, cfg.Cutlines, cfg.Ignores)
	if err := gp.Parse(cfg.Input); err != nil {
		return err
	}

	file, err := os.Create(cfg.Output)
	if err != nil {
		return fmt.Errorf("can't create %q: %v", cfg.Output, err)
	}
	defer file.Close()

	if err := gp.Report(file); err != nil {
		return err
	}

	return nil
}

// NewCLIConfig creates a new configuration based on the command-line arguments.
func NewCLIConfig() (*config.Config, error) {
	input := flag.String("i", "cover.prof", "input file name")
	output := flag.String("o", "cover.html", "output file name")
	cutlines := flag.String("cutlines", "70,40", "cutlines (safe,warning)")
	root := flag.String("root", ".", "root package name")
	ignores := flag.String("ignores", "", "ignore packages (comma separated)")
	flag.Parse()

	parsedCutlines, err := ParseCutlines(*cutlines)
	if err != nil {
		return nil, err
	}

	return &config.Config{
		Input:    *input,
		Output:   *output,
		Cutlines: parsedCutlines,
		Root:     *root,
		Ignores:  ParseIgnores(*ignores),
	}, nil
}

// ParseCutlines parses the cutlines argument.
func ParseCutlines(cutlines string) (*config.Cutlines, error) {
	frags := strings.Split(cutlines, ",")
	safe, err := strconv.ParseFloat(frags[0], 64)
	if err != nil {
		return nil, err
	}
	warning, err := strconv.ParseFloat(frags[len(frags)-1], 64)
	if err != nil {
		return nil, err
	}

	return &config.Cutlines{
		Safe:    safe,
		Warning: warning,
	}, nil
}

// ParseIgnores parses the ignores argument.
func ParseIgnores(ignores string) []string {
	if ignores == "" {
		return nil
	}
	return strings.Split(ignores, ",")
}
