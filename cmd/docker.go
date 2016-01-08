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
	"github.com/asteris-llc/vaultfs/docker"
	"github.com/asteris-llc/vaultfs/fs"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// dockerCmd represents the docker command
var dockerCmd = &cobra.Command{
	Use:   "docker {mountpoint}",
	Short: "start the docker volume server at the specified root",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("expected exactly one argument, a mountpoint")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		lockMemory()

		driver := docker.New(docker.Config{
			Root:  args[0],
			Token: viper.GetString("token"),
			Vault: fs.NewConfig(viper.GetString("address"), viper.GetBool("insecure")),
		})

		defer func() {
			for _, err := range driver.Stop() {
				logrus.WithError(err).Error("error stopping driver")
			}
		}()

		handler := volume.NewHandler(driver)
		logrus.WithField("socket", viper.GetString("socket")).Info("serving unix socket")
		err := handler.ServeUnix("root", viper.GetString("socket"))
		if err != nil {
			logrus.WithError(err).Fatal("failed serving")
		}
	},
}

func init() {
	RootCmd.AddCommand(dockerCmd)

	dockerCmd.Flags().StringP("address", "a", "https://localhost:8200", "vault address")
	dockerCmd.Flags().BoolP("insecure", "i", false, "skip SSL certificate verification")
	dockerCmd.Flags().StringP("token", "t", "", "vault token")
	dockerCmd.Flags().StringP("socket", "s", "/run/docker/plugins/vault.sock", "socket address to communicate with docker")
	viper.BindPFlags(dockerCmd.Flags())
}
