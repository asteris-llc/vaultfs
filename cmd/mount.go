// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/vaultfs/fs"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount {mountpoint}",
	Short: "mount a vault FS at the specified mountpoint",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("expected exactly one argument")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fs := fs.New()
		stop, errs := fs.Mount(args[0])

		// handle interrupt
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)

			<-c
			logrus.Info("stopping")
			stop()
		}()

		status := 0
		for err := range errs {
			status = 1
			logrus.WithError(err).Error("error mounting")
		}

		os.Exit(status)
	},
}

func init() {
	RootCmd.AddCommand(mountCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
