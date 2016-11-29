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

	"hash/crc64"
	"path"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"golang.org/x/net/context"
	"strings"
)

// These lines statically ensure that a *SnapshotsDir implement the given
// interfaces; a misplaced refactoring of the implementation that breaks
// the interface will be catched by the compiler
var _ = fs.HandleReadDirAller(&Root{})
var _ = fs.NodeStringLookuper(&Root{})

var table = crc64.MakeTable(crc64.ISO)

// Root implements both Node and Handle
type Root struct {
	root  string
	logic *api.Logical
}

// NewRoot creates a new root and returns it
func NewRoot(root string, logic *api.Logical) *Root {
	logrus.Infoln("Creating new vault root at:", root)
	return &Root{
		root:  root,
		logic: logic,
	}
}

// Attr sets attrs on the given fuse.Attr
func (Root) Attr(ctx context.Context, a *fuse.Attr) error {
	logrus.Debug("handling Root.Attr call")
	a.Inode = 0
	a.Mode = os.ModeDir | os.FileMode(0555)
	a.Uid = 0
	a.Gid = 0

	return nil
}

// Lookup looks up a path
func (r *Root) Lookup(ctx context.Context, name string) (fs.Node, error) {
	logrus.WithField("name", name).Debug("handling Root.Lookup call")

	lookupPath := path.Join(r.root, name)

	// TODO: handle context cancellation
	secret, err := r.logic.Read(lookupPath)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"root": r.root, "name": name}).Error("error reading key")
		return nil, fuse.EIO
	}

	// Literal secret
	if secret != nil {
		logrus.Debugln("Lookup succeeded for file-like secret.")
		return &Secret{
			secret,
			r.logic,
			crc64.Checksum([]byte(name), table),
			lookupPath,
		}, nil
	}


	// Not a literal secret. Try listing to see if it's a directory.
	dirSecret, err := r.logic.List(lookupPath)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"root": r.root, "name": name}).Error("error listing key")
		return nil, fuse.EIO
	}

	if dirSecret != nil {
		logrus.Debugln("Lookup succeeded for directory-like secret.")
		return &SecretDir{
			dirSecret,
			r.logic,
			crc64.Checksum([]byte(name), table),
			lookupPath,
		}, nil
	}

	logrus.Debugln("lookup failed.")
	return nil, fuse.ENOENT
}

// ReadDirAll returns a list of secrets
func (r *Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	logrus.WithField("path", r.root).Debugln("handling Root.ReadDirAll call")

	secrets, err := r.logic.List(path.Join(r.root))
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"root": r.root}).Errorln("Error listing secrets")
		return nil, fuse.EIO
	}

	if secrets.Data["keys"] == nil {
		return []fuse.Dirent{}, nil
	}

	dirs := []fuse.Dirent{}
	for i := 0; i < len(secrets.Data["keys"].([]interface{})); i++ {
		// Ensure we don't have a trailing /
		rawName := secrets.Data["keys"].([]interface{})[i].(string)
		secretName := strings.TrimRight(rawName, "/")

		inode := crc64.Checksum([]byte(secretName), table)

		var nodeType fuse.DirentType
		if strings.HasSuffix(rawName, "/") {
			nodeType = fuse.DT_Dir
		} else {
			nodeType = fuse.DT_File
		}

		d := fuse.Dirent{
			Name:  secretName,
			Inode: inode,
			Type:  nodeType,
		}
		dirs = append(dirs, d)
	}

	logrus.Debugln("ReadDirAll succeeded.")
	return dirs, nil
}
