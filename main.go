// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/goulash/util"
	flag "github.com/ogier/pflag"
)

const (
	progName    = "repoctl"
	progVersion = "0.10"
	progDate    = "1. October 2014"

	defaultRepo = "/srv/abs/atlas.db.tar.gz"
)

var defaultConfigPath = path.Join(os.Getenv("HOME"), ".repo.conf")

type IniConfig struct {
	Repo     string   `toml:"repo"`
	AddParam []string `toml:"add_params"`
	RmParam  []string `toml:"rm_params"`
}

// Config contains all the configuration flags, variables, and arguments that
// are needed for the various actions.
type Config struct {
	// ConfigFile stores the name of the configuration file from which this
	// configuration was loaded from, otherwise it is empty.
	ConfigFile string

	// Repository is the absolute path to the database. We assume that this is
	// also where the packages are. The variables database and path are constructed
	// from this.
	Repository string
	database   string
	path       string
	// AddParameters are parameters to add to the repo-add command line.
	AddParameters []string
	// RemoveParameters are parameters to add to the repo-remove command line.
	RemoveParameters []string

	// Quiet causes less information to be printed than usual.
	Quiet bool
	// Columnate causes items to be printed in columns rather than lines.
	Columnate bool

	// Versioned causes packages to be printed with version information.
	Versioned bool
	// Mode can be either "count", "filter", or "mark" (which is the default
	// if no match is found.
	Mode string
	// Pending marks packages that need to be added to the database,
	// as well as packages that are in the database but are not available.
	Pending bool
	// Duplicates marks the number of obsolete packages for each package.
	Duplicates bool
	// Installed marks whether packages are locally installed or not.
	Installed bool
	// Synchronize marks which packages have newer versions on AUR.
	Synchronize bool

	// Interactive requires confirmation before deleting and changing the
	// repository database.
	Interactive bool
	// Backup causes older packages to be backed up rather than deleted.
	// For this, the files are given the suffix ".bak".
	Backup bool

	// Arguments contains the arguments given on the commandline.
	Args []string
}

// Action is the type that all action functions need to satisfy.
type Action func(*Config) error

// actions is a map from names to action functions.
var actions map[string]Action = map[string]Action{
	"list":   List,
	"ls":     List,
	"update": Update,
	"add":    Add,
	"remove": Remove,
	"rm":     Remove,
	"status": Status,
	"filter": Filter,
	"help":   Usage,
	"usage":  Usage,
}

// Usage prints the help message for the program.
func Usage(*Config) error {
	fmt.Printf("%s %s (%s)\n", progName, progVersion, progDate)
	fmt.Print(`
Manage local pacman repositories.

Commands available:

  list             List packages that belong to the managed repository.
  ls               Options available are:
                    -v --versions   show package versions along with name
                    -d --duplicates mark packages with duplicate package files
                    -p --pending    mark pending changes to the database
                    -l --installed  mark packages that are locally installed
                    -u --outdated   mark packages that are newer in AUR
                    -a --all        same as -vpdlu

  filter <crit...> Filter list of packages by one or more criteria; each can
                   optionally be prefixed by an exclamation mark ! to negate
                   the filter.
                   Filters available are:
                    duplicates      files to be deleted or backed up
                    pending         packages to be added/removed from database
                    outdated        packages with newer versions in AUR
                    missing         packages not found in AUR
                    local           packages locally installed

  status           Show pending changes to the database and packages that can
                   be updated.

  add <pkgname>    Add the latest package(s) with <pkgname> to the database
                   and delete all obsolete package files.

  remove <pkgname> Remove the package(s) with <pkgname> from the database and
  rm               delete all the corresponding package files.

  update           Automatically scan the repository for changes and update
                   by changing the database and deleting obsolete package files.

  reset            Reset the repository database by removing it and adding all
                   up-to-date packages while deleting obsolete package files.

                   Options available to add, remove, update, and reset are:
                    -i --interactive  ask before doing anything destructive
                    -b --backup       backup obsolete package files instead of
                                      deleting; packages are put into backup/

  help             Show the usage for repoctl. Synonym for
  usage             repoctl --help

NOTE: In all of these cases, <pkgname> is the name of the package, without
anything else. For example: pacman, and not pacman-3.5.3-1-i686.pkg.tar.xz

General options available are:

 -h --help      show this usage message
 -q --quiet     only show information when absolutely necessary
 -s --columns   show items in columns rather than lines
 -c --config    configuration file to load settings from
    --repo      path to repository and database, such as
                "/srv/abs/atlas.db.tar.gz"
`)

	return nil
}

// NewConfig creates a minimal configuration.
func NewConfig(repo string) *Config {
	return &Config{
		Repository: repo,
		path:       path.Dir(repo),
		database:   path.Base(repo),
	}
}

func readIniInto(path string, conf *Config) error {
	var ini IniConfig
	_, err := toml.DecodeFile(path, &ini)
	if err != nil {
		return err
	}

	if conf.Repository == "" {
		conf.Repository = ini.Repo
	}
	conf.AddParameters = ini.AddParam
	conf.RemoveParameters = ini.RmParam

	return nil
}

// ReadConfig reads a configuration from the command line arguments.
func ReadConfig() (conf *Config, cmd Action, err error) {
	var allListOptions bool
	var showHelp bool
	conf = &Config{}

	flag.StringVarP(&conf.ConfigFile, "config", "c", defaultConfigPath, "configuration file to load settings from")
	flag.StringVar(&conf.Repository, "repo", "", "path to repository and database")

	flag.BoolVarP(&conf.Columnate, "columns", "s", false, "show items in columns rather than lines")
	flag.BoolVar(&conf.Quiet, "quiet", false, "show minimal amount of information")
	flag.BoolVarP(&showHelp, "help", "h", false, "show this usage message")

	// List options
	flag.BoolVarP(&conf.Versioned, "versioned", "v", false, "show package versions along with name")
	flag.BoolVarP(&conf.Pending, "pending", "p", false, "mark pending changes to the database")
	flag.BoolVarP(&conf.Duplicates, "duplicates", "d", false, "mark packages with duplicate package files")
	flag.BoolVarP(&conf.Installed, "installed", "l", false, "mark packages that are locally installed")
	flag.BoolVarP(&conf.Synchronize, "outdated", "u", false, "mark packages that are newer in AUR")
	flag.BoolVarP(&allListOptions, "all", "a", false, "all information; same as -vpdlo")

	flag.BoolVarP(&conf.Interactive, "interactive", "i", false, "ask before doing anything destructive")
	flag.BoolVarP(&conf.Backup, "backup", "b", false, "backup obsolete package files instead of deleting")

	flag.Usage = func() { Usage(nil) }
	flag.Parse()

	if showHelp {
		return nil, Usage, nil
	}

	// Reading config file and constructing path and database parts
	if ex, _ := util.FileExists(conf.ConfigFile); ex {
		err := readIniInto(conf.ConfigFile, conf)
		if err != nil {
			return nil, nil, err
		}
	} else {
		fmt.Fprintf(os.Stderr, "Warning: missing config file %q.\n", conf.ConfigFile)
	}

	if conf.Repository == "" {
		fmt.Fprintf(os.Stderr, "Warning: missing repository, using %q.\n", defaultRepo)
		conf.Repository = defaultRepo
	}
	conf.path = path.Dir(conf.Repository)
	conf.database = path.Base(conf.Repository)
	if len(flag.Args()) == 0 {
		return nil, Usage, errors.New("no action specified on command line")
	}

	if allListOptions {
		conf.Versioned = true
		conf.Pending = true
		conf.Duplicates = true
		conf.Installed = true
		conf.Synchronize = true
	}

	conf.Args = flag.Args()[1:]
	cmd, ok := actions[flag.Arg(0)]
	if !ok {
		return nil, Usage, errors.New("unrecognized action " + flag.Arg(0))
	}

	return conf, cmd, nil
}

func (c *Config) inform(v interface{}) {
	if !c.Quiet {
		if e, ok := v.(error); ok {
			fmt.Fprintf(os.Stderr, "warning: %s\n", e)
		} else {
			fmt.Fprintln(os.Stderr, v)
		}
	}
}

func main() {
	conf, cmd, err := ReadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println()
		Usage(nil)
		os.Exit(1)
	}

	cmd(conf)
}
