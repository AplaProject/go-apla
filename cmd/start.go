// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package cmd

import (
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/daylight"
	"github.com/spf13/cobra"
)

// startCmd is starting node
var startCmd = &cobra.Command{
	Use:    "start",
	Short:  "Starting node",
	PreRun: loadConfigWKey,
	Run: func(cmd *cobra.Command, args []string) {
		daylight.Start()
	},
}

func init() {
	time.Local = time.UTC
	startCmd.Flags().BoolVar(&conf.Config.TestRollBack, "testRollBack", false, "Starts special set of daemons")
	startCmd.Flags().BoolVar(&conf.Config.FuncBench, "funcBench", false, "Disable access checking in some built-in functions for benchmarks")
}
