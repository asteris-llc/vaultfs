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
	"os"
	"os/signal"

	"crypto/tls"
	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/vaultfs/fs"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
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
		config := &api.Config{
			Address: viper.GetString("address"),
			HttpClient: &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: viper.GetBool("insecure"),
					},
				},
			},
		}

		logrus.WithField("address", viper.GetString("address")).Info("creating FUSE client for Vault")
		fs, err := fs.New(config, viper.GetString("token"), viper.GetString("root"))
		if err != nil {
			logrus.WithError(err).Fatal("error creatinging fs")
		}

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

	mountCmd.Flags().StringP("root", "r", "secret", "root path for reads")
	mountCmd.Flags().StringP("address", "a", "https://localhost:8200", "vault address")
	mountCmd.Flags().BoolP("insecure", "i", false, "skip SSL certificate verification")
	mountCmd.Flags().StringP("token", "t", "", "vault token")
	viper.BindPFlags(mountCmd.Flags())
}
