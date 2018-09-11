package dshRepIO

import (
	"io"
	"strings"
	"regexp"
	"path/filepath"
	"os"
	"../utils"
)

type URI interface {
	Protocol() string
	Path() string
	ToString() string
}

func URIFromStr(s string) (URI, error) {
	sp_s := strings.SplitN(s, "://", 2)
	switch strings.ToLower(sp_s[0]) {
	case "file":
		return newFileURI(sp_s[1])
	default:
		return newFileURI(s)
	}
}

func GetURI(uri interface{}) (URI, error) {
	if _u, ok := uri.(URI); ok {
		return _u, nil
	} else if _u, ok := uri.(string); ok {
		return URIFromStr(_u)
	}
	return nil, utils.Error("Invalid uri type!")
}

type FileURI struct {
	path string
}

var windowsAbsPathWithSlash = regexp.MustCompile(`^[\\/]\w:([\\/].*?)*`)
var windowsAbsPath = regexp.MustCompile(`^\w:([\\/].*?)*`)

func newFileURI(s string) (*FileURI, error) {
	p := s
	if windowsAbsPathWithSlash.MatchString(p) {
		p = s[1:]
	}
	path, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}
	return &FileURI{path:path}, nil
}

func (u *FileURI) Protocol() string {
	return "file"
}

func (u *FileURI) Path() string {
	return u.path
}

func (u *FileURI) ToString() string {
	if windowsAbsPath.MatchString(u.path) {
		return "file:///" + filepath.ToSlash(u.path)
	}
	return  "file://" + filepath.ToSlash(u.path)
}

type AccessType int

const (
	AccessFileSystem AccessType = iota
)

type Accessor interface {
	io.Reader
	io.Seeker
	io.Writer
	Type() AccessType
	URI() URI
	Comment()
}

type FSAccessor struct {
	*os.File
	uri URI
	isOpen bool
}

func NewFSAccessor(uri URI) (*FSAccessor, error) {
	if uri.Protocol() != "file" {
		return nil, utils.Error("The URI protocol of a FSAccessor need to be 'file'")
	}
	return &FSAccessor{uri:uri, isOpen:false}, nil
}

func (a *FSAccessor) Open(flag int, perm os.FileMode) error {
	if a.isOpen {
		return nil
	}
	f, err := os.OpenFile(a.uri.Path(), flag, perm)
	a.File = f
	if err == nil {
		a.isOpen = true
	}
	return err
}

func (a *FSAccessor) Close() error {
	if !a.isOpen {
		return nil
	}
	err := a.File.Close()
	if err == nil {
		a.isOpen = false
		a.File = nil
	}
	return err
}

func (a *FSAccessor) Type() AccessType {
	return AccessFileSystem
}

func (a *FSAccessor) URI() URI {
	return a.uri
}

func (a *FSAccessor) Comment()  {

}
