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
	"github.com/hashicorp/vault/api"
	"golang.org/x/net/context"
	"hash/crc64"
	"path"
)

var table = crc64.MakeTable(crc64.ISO)

// Root implements both Node and Handle
type Root struct {
	root  string
	logic *api.Logical
}

// NewRoot creates a new root and returns it
func NewRoot(root string, logic *api.Logical) *Root {
	return &Root{
		root:  root,
		logic: logic,
	}
}

// Attr sets attrs on the given fuse.Attr
func (Root) Attr(ctx context.Context, a *fuse.Attr) error {
	logrus.Debug("handling Root.Attr call")
	a.Inode = 1
	a.Mode = os.ModeDir
	return nil
}

// Lookup looks up a path
func (r *Root) Lookup(ctx context.Context, name string) (fs.Node, error) {
	logrus.WithField("name", name).Debug("handling Root.Lookup call")

	// TODO: handle context cancellation
	secret, err := r.logic.Read(path.Join(r.root, name))
	if secret == nil && err == nil {
		return nil, fuse.ENOENT
	} else if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"root": r.root, "name": name}).Error("error reading key")
		return nil, fuse.EIO
	}

	return Secret{
		secret,
		crc64.Checksum([]byte(name), table),
	}, nil
}

// ReadDirAll returns a list of secrets
func (r *Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	logrus.Debug("handling Root.ReadDirAll call")

	secrets, err := r.logic.List(path.Join(r.root))
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"root": r.root}).Error("error reading secrets")
	}
	if secrets.Data["keys"] == nil {
		return []fuse.Dirent{}, nil
	}

	dirs := []fuse.Dirent{}
	for i := 0; i < len(secrets.Data["keys"].([]interface{})); i++ {
		d := fuse.Dirent{
			Name:  secrets.Data["keys"].([]interface{})[i].(string),
			Inode: 1,
			Type:  fuse.DT_File, // TODO: A lie, consider an alternative
		}
		dirs = append(dirs, d)
	}

	return dirs, nil
}
