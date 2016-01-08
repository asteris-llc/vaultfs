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
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
	"os"
	"path"
	"sync"
)

// Driver implements the interface for a Docker volume plugin
type Driver struct {
	config  Config
	servers map[string]*Server
	m       *sync.Mutex
}

// New instantiates a new driver and returns it
func New(config Config) Driver {
	return Driver{
		config:  config,
		servers: map[string]*Server{},
		m:       new(sync.Mutex),
	}
}

// Create handles volume creation calls
func (d Driver) Create(r volume.Request) volume.Response {
	return volume.Response{}
}

// Remove handles volume removal calls
func (d Driver) Remove(r volume.Request) volume.Response {
	d.m.Lock()
	defer d.m.Unlock()
	mount := d.mountpoint(r.Name)
	logger := logrus.WithFields(logrus.Fields{
		"name":       r.Name,
		"mountpoint": mount,
	})
	logger.Debug("got remove request")

	if server, ok := d.servers[mount]; ok {
		if server.connections <= 1 {
			logger.Debug("removing server")
			delete(d.servers, mount)
		}
	}

	return volume.Response{}
}

// Path handles calls for mountpoints
func (d Driver) Path(r volume.Request) volume.Response {
	return volume.Response{Mountpoint: d.mountpoint(r.Name)}
}

// Mount handles creating and mounting servers
func (d Driver) Mount(r volume.Request) volume.Response {
	d.m.Lock()
	defer d.m.Unlock()

	mount := d.mountpoint(r.Name)
	logger := logrus.WithFields(logrus.Fields{
		"name":       r.Name,
		"mountpoint": mount,
	})
	logger.Info("mounting volume")

	server, ok := d.servers[mount]
	if ok && server.connections > 0 {
		server.connections++
		return volume.Response{Mountpoint: mount}
	}

	mountInfo, err := os.Lstat(mount)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(mount, 0444); err != nil {
			logger.WithError(err).Error("error making mount directory")
			return volume.Response{Err: err.Error()}
		}
	} else if err != nil {
		logger.WithError(err).Error("error checking if directory exists")
		return volume.Response{Err: err.Error()}
	}

	if mountInfo != nil && !mountInfo.IsDir() {
		logger.Error("already exists and not a directory")
		return volume.Response{Err: fmt.Sprintf("%s already exists and is not a directory", mount)}
	}

	server, err = NewServer(d.config.Vault, mount, d.config.Token, r.Name)
	if err != nil {
		logger.WithError(err).Error("error creating server")
		return volume.Response{Err: err.Error()}
	}

	go server.Mount()
	d.servers[mount] = server

	return volume.Response{Mountpoint: mount}
}

// Unmount handles unmounting (but not removing) servers
func (d Driver) Unmount(r volume.Request) volume.Response {
	d.m.Lock()
	defer d.m.Unlock()

	mount := d.mountpoint(r.Name)
	logger := logrus.WithFields(logrus.Fields{
		"name":       r.Name,
		"mountpoint": mount,
	})
	logger.Info("unmounting volume")

	if server, ok := d.servers[mount]; ok {
		logger.WithField("conns", server.connections).Debug("found server")
		if server.connections == 1 {
			logger.Debug("unmounting")
			err := server.Unmount()
			if err != nil {
				logger.WithError(err).Error("error unmounting server")
				return volume.Response{Err: err.Error()}
			}
			server.connections--
		}
	} else {
		logger.Error("could not find volume")
		return volume.Response{Err: fmt.Sprintf("unable to find the volume mounted at %s", mount)}
	}

	return volume.Response{}
}

func (d Driver) mountpoint(name string) string {
	return path.Join(d.config.Root, name)
}

// Stop stops all the servers
func (d Driver) Stop() []error {
	d.m.Lock()
	defer d.m.Unlock()
	logrus.Debug("got stop request")

	errs := []error{}
	for _, server := range d.servers {
		err := server.Unmount()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
