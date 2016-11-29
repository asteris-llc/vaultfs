package fs

import (
	"bazil.org/fuse"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"golang.org/x/net/context"
	"os"
	"bazil.org/fuse/fs"
	"path"
	"hash/crc64"
	"strings"
)

// Statically ensure that *dir implement those interface
var _ = fs.HandleReadDirAller(&SecretDir{})
var _ = fs.NodeStringLookuper(&SecretDir{})

// SecretDir implements Node and Handle
// This is the type we return if the Secret is a secret that we only were able to get via a list - i.e. is directory-like.
type SecretDir struct {
	*api.Secret
	logic *api.Logical
	inode uint64
	lookupPath string
}

// Attr returns attributes about this Secret
func (s SecretDir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = s.inode
	a.Mode = os.ModeDir | 0555
	a.Uid = 0
	a.Gid = 0

	return nil
}

// Lookup looks up a path
func (s *SecretDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	logrus.WithField("name", name).Debug("handling Root.Lookup call")

	lookupPath := path.Join(s.lookupPath, name)

	// TODO: handle context cancellation
	secret, err := s.logic.Read(lookupPath)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"root": s.lookupPath, "name": name}).Error("error reading key")
		return nil, fuse.EIO
	}

	// Literal secret
	if secret != nil {
		logrus.Debugln("Lookup succeeded for file-like secret.")
		return &Secret{
			secret,
			s.logic,
			crc64.Checksum([]byte(name), table),
			lookupPath,
		}, nil
	}


	// Not a literal secret. Try listing to see if it's a directory.
	dirSecret, err := s.logic.List(lookupPath)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"root": s.lookupPath, "name": name}).Error("error listing key")
		return nil, fuse.EIO
	}

	if secret != nil {
		logrus.Debugln("Lookup succeeded for directory-like secret.")
		return &SecretDir{
			dirSecret,
			s.logic,
			crc64.Checksum([]byte(name), table),
			lookupPath,
		}, nil
	}

	logrus.Debugln("lookup failed.")
	return nil, fuse.ENOENT
}

// ReadDirAll returns a list of secrets in this directory
func (s *SecretDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	logrus.WithField("path", s.lookupPath).Debugln("handling Root.ReadDirAll call")

	if s.Data["keys"] == nil {
		return []fuse.Dirent{}, nil
	}

	dirs := []fuse.Dirent{}
	for i := 0; i < len(s.Data["keys"].([]interface{})); i++ {
		// Ensure we don't have a trailing /
		rawName := s.Data["keys"].([]interface{})[i].(string)
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

	logrus.Debugln("ReadDirAll succeeded.", dirs)
	return dirs, nil
}