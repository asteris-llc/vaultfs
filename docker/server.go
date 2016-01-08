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

package docker

import (
	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/vaultfs/fs"
	"github.com/hashicorp/vault/api"
)

// Server wraps VaultFS and tracks connection counts
type Server struct {
	fs          *fs.VaultFS
	connections int
	stopFunc    func()
	errs        chan error
}

// NewServer returns a new server with initial state
func NewServer(config *api.Config, mountpoint, token, root string) (*Server, error) {
	fs, err := fs.New(config, mountpoint, token, root)
	if err != nil {
		return nil, err
	}

	return &Server{fs: fs, connections: 1}, nil
}

// Mount mounts the wrapped FS on a given mountpoint. It also starts watching
// for errors, which it will log.
func (s *Server) Mount() error {
	err := s.fs.Mount()

	if err != nil {
		logrus.WithError(err).Error("error in server, stopping")
		return err
	}

	return nil
}

// Unmount stops the wrapped FS. It returns the last error that it sees, but
// will log any others it receives.
func (s *Server) Unmount() error {
	err := s.fs.Unmount()

	if err != nil {
		logrus.WithError(err).Error("could not unmount cleanly")
	}

	return err
}
