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
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

// Secret implements Node and Handle
type Secret struct{}

const greeting = "hello, world\n"

// Attr returns attributes about this Secret
func (Secret) Attr(ctx context.Context, a *fuse.Attr) error {
	logrus.Debug("handling Secret.Attr call")
	a.Inode = 2
	a.Mode = 0444
	a.Size = uint64(len(greeting))
	return nil
}

// ReadAll gets the content of this Secret
func (Secret) ReadAll(ctx context.Context) ([]byte, error) {
	logrus.Debug("handling Secret.ReadAll call")
	return []byte(greeting), nil
}
