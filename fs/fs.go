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

package fs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/Sirupsen/logrus"
)

// VaultFS is a vault filesystem
type VaultFS struct{}

// New returns a new VaultFS
func New() *VaultFS {
	logrus.Debug("created new FS")
	return &VaultFS{}
}

// Mount the FS at the given mountpoint
func (v *VaultFS) Mount(mountpoint string) error {
	conn, err := fuse.Mount(
		mountpoint,
		fuse.FSName("vault"),
		fuse.VolumeName("vault"),
	)
	logrus.Debug("created conn")
	if err != nil {
		return err
	}
	defer conn.Close()

	logrus.Debug("starting to serve")
	err = fs.Serve(conn, v)
	if err != nil {
		return err
	}

	<-conn.Ready
	return conn.MountError
}

// Root returns the struct that does the actual work
func (VaultFS) Root() (fs.Node, error) {
	logrus.Debug("returning root")
	return Root{}, nil
}
