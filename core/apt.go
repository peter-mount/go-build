package core

import (
	"fmt"
	"github.com/peter-mount/go-build/util"
	"github.com/peter-mount/go-build/util/arch"
	"github.com/peter-mount/go-build/util/makefile/target"
	"github.com/peter-mount/go-build/util/meta"
	"github.com/peter-mount/go-kernel/v2/util/walk"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
)

type Apt struct {
	Encoder *Encoder `kernel:"inject"`
	Build   *Build   `kernel:"inject"`
	Apt     *string  `kernel:"flag,apt,apt archive to generate"`
	AptSrc  *string  `kernel:"flag,apt-src,source from build"`
	config  Config
}

type Config struct {
	Package Package `yaml:"package"`
	Lintian bool    `yaml:"lintian"`
}

type Package struct {
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"`
	Release       string   `yaml:"release"`
	Maintainer    string   `yaml:"maintainer"`
	Homepage      string   `yaml:"homepage"`
	Description   string   `yaml:"description"`
	Architectures []string `yaml:"architecture"`
	Depends       []string `yaml:"depends"`
}

func (s *Apt) Start() error {
	if err := s.loadConfig(); err != nil {
		// Ignore if debian.yaml does not exist
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	s.Build.AddExtension(s.extension)

	if *s.Apt != "" {
		return s.run()
	}
	return nil
}

func (s *Apt) extension(arch arch.Arch, target target.Builder, meta *meta.Meta) {

	// Filter to only supported platforms
	if len(s.config.Package.Architectures) != 0 {
		platform := arch.Platform()
		found := false
		for _, a := range s.config.Package.Architectures {
			found = found || a == platform
		}
		if !found {
			return
		}
	}

	// Apt package to generate
	aptName := s.config.Package.AptName(arch.Arch())
	destDir := filepath.Join(*s.Encoder.Dest, "apt", aptName)
	debName := destDir + ".deb"

	// Generate copy for deployment
	target.
		Target(debName, arch.Target()+"_dist").
		//MkDir(destDir).
		//Echo("INSTALL", destDir).
		//BuildTool("-copydir", arch.BaseDir(*s.Encoder.Dest), "-d", filepath.Join(destDir, "usr/local", config.Package.Name)).
		Echo("DIST APT", debName).
		BuildTool(
			"-apt", debName,
			"-apt-src", arch.BaseDir(*s.Encoder.Dest),
			"-build-platform", arch.Platform(),
			"-d", destDir)

	// Add apt rule which depends on dist & the deb file(s)
	meta.ArchTarget.Rule(arch.Target()+"_apt", arch.Target()+"_dist", debName)
}

func (s *Apt) run() error {
	if *s.Encoder.Dest == "" {
		panic("-d required for apt!")
	}
	if *s.AptSrc == "" {
		panic("-apt-src required for apt!")
	}

	err := s.copyDist()

	if err == nil {
		err = s.installControl()
	}

	if err == nil {
		err = s.dpkg()
	}

	return err
}

func (p Package) AptName(arch string) string {
	return fmt.Sprintf("%s_%s-%s_%s", p.Name, p.Version, p.Release, arch)
}

func (s *Apt) installControl() error {
	p := s.config.Package
	c := fmt.Sprintf("Package: %s\nVersion: %s\nMaintainer: %s\nHomepage: %s\nDescription: %s\n",
		p.Name, p.Version, p.Maintainer, p.Homepage, p.Description)
	if len(p.Depends) > 0 {
		c = c + "Depends: " + strings.Join(p.Depends, " ") + "\n"
	}

	a := strings.Split(*s.Build.Platforms, ":")
	c = c + "Architecture: " + a[1] + a[2] + "\n"

	fName := filepath.Join(*s.Encoder.Dest, "DEBIAN", "control")
	err := os.MkdirAll(filepath.Dir(fName), 0755)
	if err == nil {
		err = os.WriteFile(fName, []byte(c), 0644)
	}
	return err
}

func (s *Apt) loadConfig() error {
	b, err := os.ReadFile("debian.yaml")
	if err == nil {
		err = yaml.Unmarshal(b, &s.config)
	}

	return err
}

func (s *Apt) copyDist() error {
	// create base package directory
	_ = os.RemoveAll(*s.Encoder.Dest)
	return walk.NewPathWalker().
		Then(s.copyDir).
		Walk(*s.AptSrc)
}

func (s *Apt) copyDir(path string, info os.FileInfo) error {
	// Ignore the source base directory
	if path == *s.Encoder.Dest {
		return nil
	}

	// dest is the source minus the source base directory name
	dstName := filepath.Join(*s.Encoder.Dest, "/usr/local/", s.config.Package.Name, strings.TrimPrefix(path, *s.AptSrc))

	if info.IsDir() {
		return os.MkdirAll(dstName, info.Mode())
	}

	return util.CopyFile(path, dstName, info)
}

func (s *Apt) dpkg() error {
	err := util.RunCommand("dpkg", "--build", *s.Encoder.Dest, *s.Apt)

	if err == nil && s.config.Lintian {
		util.Label("LINTIAN", *s.Apt)
		err = util.RunCommand("lintian", *s.Apt)
	}

	return err
}
