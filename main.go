// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Command interface {
	Name() string
	Parse(args []string) error
	Exec()

	RepoDir() string
	Path(args ...string) string
}

type Common struct {
	Root    string
	RootAbs string

	Package string
	Repo    string
}

func (cmd Common) RepoDir() string {
	return cmd.Path("src", filepath.FromSlash(cmd.Package))
}

func (cmd Common) Path(args ...string) string {
	return filepath.Join(append([]string{cmd.RootAbs}, args...)...)
}

func main() {
	common := Common{}
	cmdname := ""
	args := []string{}
	{
		set := flag.NewFlagSet("", flag.ContinueOnError)

		set.StringVar(&common.Root, "root", os.Getenv("GOSPACE_ROOT"), "root directory (default GOSPACE_ROOT)")
		set.StringVar(&common.Package, "pkg", os.Getenv("GOSPACE_PKG"), "package name (default GOSPACE_PKG)")
		set.StringVar(&common.Repo, "repo", os.Getenv("GOSPACE_REPO"), "package name (default GOSPACE_REPO)")

		if err := set.Parse(os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "invalid args")
			os.Exit(1)
		}

		fail := false
		if common.Root == "" {
			fmt.Fprintln(os.Stderr, "root directory is missing, please specify `-root` or GOSPACE_ROOT environment variable")
			fail = true
		}
		if common.Package == "" {
			fmt.Fprintln(os.Stderr, "package name is missing, please specify `-pkg` or GOSPACE_PKG environment variable")
			fail = true
		}
		if common.Repo == "" {
			fmt.Fprintln(os.Stderr, "repo name is missing, please specify `-repo` or GOSPACE_REPO environment variable")
			fail = true
		}

		if fail {
			os.Exit(1)
		}

		cmdname = set.Arg(0)
		if set.NArg() > 1 {
			args = set.Args()[1:]
		}

		common.RootAbs, _ = filepath.Abs(common.Root)
	}

	cmds := []Command{
		&Setup{Common: common},
		&Update{Common: common},
		// &Cache{Common: common},

		&IsTidy{Common: common},
		&Hash{Common: common},
		&ZipVendor{Common: common},
		&UnzipVendor{Common: common},
		&FlattenVendor{Common: common},
	}

	for _, cmd := range cmds {
		if strings.EqualFold(cmdname, cmd.Name()) {
			if err := cmd.Parse(args); err != nil {
				fmt.Fprintln(os.Stderr, "invalid args", err)
				os.Exit(1)
			}
			Exec(cmd)
			return
		}
	}

	fmt.Fprintln(os.Stderr, "unknown command:", cmdname)
	fmt.Fprintln(os.Stderr, "supported:")
	for _, cmd := range cmds {
		fmt.Fprintln(os.Stderr, "\t"+cmd.Name())
	}
	os.Exit(1)
}

func Exec(cmd Command) {
	gomodfilename := filepath.Join(cmd.RepoDir(), "go.mod")

	gomod, gomoderr := ioutil.ReadFile(gomodfilename)
	defer func() {
		if gomoderr != nil {
			return
		}
		gomodnew, gomodnewerr := ioutil.ReadFile(gomodfilename)
		if gomodnewerr == nil && !bytes.Equal(gomod, gomodnew) {
			ioutil.WriteFile(gomodfilename, gomod, 0644)
		}
	}()

	cmd.Exec()
}
