// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/spf13/cobra"
)

// Reset -------------------------------------------------------------

var ResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "recreate repository database",
	Long: `Delete the repository database and re-add all packages in repository.
    
  Essentially, this command deletes the repository database and
  recreates it by running the update command.
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := repo.Reset(repo.PrinterEH(os.Stderr))
		dieOnError(err)
	},
}

// Add ---------------------------------------------------------------

var movePackages bool

func init() {
	AddCmd.Flags().BoolVarP(&movePackages, "move", "m", false, "move packages into repository")
}

var AddCmd = &cobra.Command{
	Use:   "add <pkgfile...>",
	Short: "copy and add packages to the repository",
	Long: `Add (and copy if necessary) the package file to the repository.

  Similarly to the repo-add script, this command copies the package
  file to the repository (if not already there) and adds the package to
  the database.  Exactly this package is added to the database, this
  allows you to downgrade a package in the repository.

  Any other package files in the repository are deleted or backed up,
  depending on whether the backup option is given. If the backup option
  is given, the "obsolete" package files are moved to a backup
  directory of choice.

  Note: since version 0.14, the semantic meaning of this command has
        changed. See the update command for the old behavior.
`,
	Example: `  repoctl add ./fairsplit-1.0.pkg.tar.gz`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if movePackages {
			err = Repo.MoveAll(args, repo.PrinterEH(os.Stderr))
		} else {
			err = Repo.AddAll(args, repo.PrinterEH(os.Stderr))
		}
		dieOnError(err)
	},
}

// Remove ------------------------------------------------------------

var RemoveCmd = &cobra.Command{
	Use:     "remove <pkgname...>",
	Aliases: []string{"rm"},
	Short:   "remove and delete packages from the database",
	Long: `Remove and delete the package files from the repository.

  This command removes the specified package from the repository
  database, and deletes any associated package files, unless the backup
  option is given, in which case the package files are moved to the
  backup directory.
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := repo.Remove(args, repo.PrinterEH)
		dieOnError(err)
	},
}