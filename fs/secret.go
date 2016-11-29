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
	"encoding/json"
	"os"

	"bazil.org/fuse"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"golang.org/x/net/context"
	"bazil.org/fuse/fs"
"io"
)

// Statically ensure that *file implements the given interface
var _ = fs.HandleReader(&Secret{})
var _ = fs.HandleReleaser(&Secret{})

// Secret implements Node and Handle
type Secret struct {
	*api.Secret
	logic *api.Logical
	inode uint64
	lookupPath string
}

func (s Secret) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return nil
}

// Attr returns attributes about this Secret
func (s Secret) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = s.inode
	a.Mode = os.FileMode(0444)

	content, err := s.ReadAll(ctx)
	if err != nil {
		logrus.WithError(err).Error("could not determine content length")
		return fuse.EIO
	}

	a.Size = uint64(len(content))
	return nil
}

func (s Secret) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data, err := s.ReadAll(ctx)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}
	resp.Data = data[:len(data)]
	return err
}

// ReadAll gets the content of this Secret
func (s Secret) ReadAll(ctx context.Context) ([]byte, error) {
	data, err := json.Marshal(s)
	if err != nil {
		logrus.Errorln("Error marshalling secret:", err)
	}
	return data, err
}

