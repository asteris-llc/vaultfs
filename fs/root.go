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
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

// Root implements both Node and Handle
type Root struct{}

// Attr sets attrs on the given fuse.Attr
func (Root) Attr(ctx context.Context, a *fuse.Attr) error {
	logrus.Debug("handling Root.Attr call")
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

// Lookup looks up a path
func (Root) Lookup(ctx context.Context, name string) (fs.Node, error) {
	logrus.Debug("handling Root.Lookup call")
	if name == "hello" {
		return Secret{}, nil
	}

	return nil, fuse.ENOENT
}

// ReadDirAll returns nothing, as Vault doesn't allow listing keys
func (Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	logrus.Debug("handling Root.ReadDirAll call")
	return []fuse.Dirent{}, fuse.EPERM
}
