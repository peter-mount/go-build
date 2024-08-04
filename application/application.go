package application

import (
	"github.com/peter-mount/go-build/version"
	"os"
	"path/filepath"
)

type FileType int

const (
	// BIN represents the path to a compiled binary
	BIN FileType = iota
	// CONFIG represents the path to configuration files
	CONFIG
	// STATIC represents the path to static files, e.g. web pages, templates or precompiled but static data
	STATIC
	// DATA represents the path to data files, e.g. databases
	DATA
	// CACHE represents the path to cached data
	CACHE
)

// AppName returns the currently running command name without any path to said binary
func AppName() string {
	return appName
}

// FileName returns the path to a required file based on it's FileType
func FileName(fileType FileType, name ...string) string {
	var dir string
	switch fileType {
	case BIN:
		dir = binDir
	case CONFIG:
		dir = etcDir
	case STATIC:
		dir = staticDir
	case DATA:
		dir = dataDir
	case CACHE:
		dir = cacheDir
	default:
	}

	return filepath.Join(dir, filepath.Join(name...))
}

var (
	appName   string
	binDir    string
	etcDir    string
	staticDir string
	dataDir   string
	cacheDir  string
)

func init() {
	arg0 := os.Args[0]
	appName = filepath.Base(arg0)
	binDir = filepath.Dir(arg0)

	switch binDir {
	// Linux FHS2.3 layout
	case "/bin", "/usr/bin", "/usr/sbin":
		projName := version.Application
		if projName == "" {
			projName = appName
		}

		etcDir = "/etc/" + projName
		staticDir = "/usr/share/" + projName
		dataDir = "/var/lib/" + projName
		cacheDir = "/var/cache/" + projName

	// Linux FHS2.3 layout under /usr/local
	case "/usr/local/bin":
		projName := version.Application
		if projName == "" {
			projName = appName
		}

		etcDir = "/usr/local/etc/" + projName
		staticDir = "/usr/local/share/" + projName
		dataDir = "/var/local/lib/" + projName
		cacheDir = "/var/local/cache/" + projName

	// Default flat layout
	default:
		rootDir := filepath.Dir(binDir)
		etcDir = rootDir + "/etc"
		staticDir = rootDir + "/share"
		dataDir = rootDir + "/data"
		cacheDir = rootDir + "/cache"
	}
}
