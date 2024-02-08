package core

import (
	"errors"
	"fmt"
	"github.com/peter-mount/go-build/util/arch"
	"github.com/peter-mount/go-build/util/jenkinsfile"
	"github.com/peter-mount/go-build/util/makefile"
	"github.com/peter-mount/go-build/util/makefile/target"
	"github.com/peter-mount/go-build/util/meta"
	"github.com/peter-mount/go-build/version"
	"github.com/peter-mount/go-kernel/v2/log"
	"github.com/peter-mount/go-kernel/v2/util/walk"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Build struct {
	Encoder          *Encoder         `kernel:"inject"`
	Dest             *string          `kernel:"flag,build,generate build files"`
	Platforms        *string          `kernel:"flag,build-platform,platform(s) to build"`
	Dist             *string          `kernel:"flag,dist,distribution destination"`
	BlockList        *string          `kernel:"flag,block,block list"`
	BuildNode        *string          `kernel:"flag,build-node,Jenkins node to run on,go"`
	Parallelize      *bool            `kernel:"flag,build-parallel,parallelize Jenkinsfile"`
	ArchiveArtifacts *string          `kernel:"flag,build-archiveArtifacts,archive files on completion"`
	NoTools          *bool            `kernel:"flag,build-no-tools,set if no tools are defined"`
	BuildLocal       *bool            `kernel:"flag,build-local,Build for local platform only"`
	libProviders     []LibProvider    // Deprecated
	extensions       Extension        // Extensions to run
	documentation    Documentation    // Documentation extensions to run
	cleanDirectories sort.StringSlice // Directories to clean other than builds and dist
	buildArch        arch.Arch        // The build platform architecture
}

// LibProvider handles calls to generate additional files/directories in a build
// returns destPath and arguments to pass
// Deprecated
type LibProvider func(builds string) (string, []string)

// AddLibProvider Deprecated
func (s *Build) AddLibProvider(p LibProvider) {
	s.libProviders = append(s.libProviders, p)
}

func (s *Build) AddCleanDirectory(dir string) {
	for _, d := range s.cleanDirectories {
		if d == dir {
			return
		}
	}
	s.cleanDirectories = append(s.cleanDirectories, dir)
	s.cleanDirectories.Sort()
}

func (s *Build) Start() error {
	// Set the build architecture
	s.buildArch = arch.Arch{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	}

	// Set the clean directory list to include our defaults
	s.AddCleanDirectory(*s.Encoder.Dest)
	s.AddCleanDirectory(*s.Dist)
	if *s.BlockList != "" {
		if err := arch.LoadBlockList(*s.BlockList); err != nil {
			return err
		}
	}

	return nil
}

func (s *Build) Run() error {
	if log.IsVerbose() {
		log.Println(version.Version)
	}

	if *s.Dest != "" {
		meta, err := meta.New()
		if err != nil {
			return err
		}

		arch, err := arch.GetArches()
		if err != nil {
			return err
		}

		tools, err := s.getTools()
		if err != nil {
			return err
		}

		err = s.generate(tools, arch, meta)
		if err != nil {
			return err
		}

		err = s.platformIndex(arch)
		if err != nil {
			return err
		}

		return s.jenkinsfile(arch)
	}
	return nil
}

func (s *Build) getTools() ([]string, error) {
	var tools []string

	if err := walk.NewPathWalker().
		Then(func(path string, info os.FileInfo) error {
			if info.Name() == "main.go" {
				tool := filepath.Base(filepath.Dir(filepath.Dir(path)))
				if tool != "build" {
					tools = append(tools, tool)
				}
			}
			return nil
		}).
		IsFile().
		Walk("tools"); err != nil {
		return nil, err
	}

	if len(tools) == 0 && !*s.NoTools {
		return nil, errors.New("no tools to compile")
	}

	sort.SliceStable(tools, func(i, j int) bool {
		return tools[i] < tools[j]
	})

	return tools, nil
}

func (s *Build) generate(tools []string, arches []arch.Arch, meta *meta.Meta) error {

	builder := makefile.New()
	builder.Comment("Generated Makefile %s", meta.Time).
		SetVar("BUILD", meta.ToolName).
		SetVar("export BUILD_VERSION", "%q", meta.Version).
		SetVar("export BUILD_TIME", "%q", meta.Time).
		SetVar("export BUILD_PACKAGE_NAME", "%q", meta.PackageName).
		SetVar("export BUILD_PACKAGE_PREFIX", "%q", meta.PackagePrefix).
		Phony("all", "clean", "init", "test")

	s.init(builder)
	s.clean(builder)
	s.test(builder)

	root, platforms := s.allRule(arches, builder)
	allPlatforms := len(platforms) == 0

	targetGroups := s.targetGroups(arches, root, platforms)

	// Used for name searching
	rootTarget := target.New()

	for _, arch := range arches {
		archTarget := targetGroups.Get(arch.Target())
		// Generate targets if all platforms or just the requested ones
		if allPlatforms || platforms[arch.Target()] {

			// Put all tools under their own target
			toolTarget := archTarget.Rule(arch.Target() + "_tools")
			for _, tool := range tools {
				if !arch.IsToolBlocked(tool) {
					s.goBuild(arch, toolTarget, tool, meta)
				}
			}

			// Apply extensions
			extTarget := archTarget.Rule(arch.Target() + "_ext")
			targetBuilder := rootTarget.New()
			s.extensions.Do(arch, targetBuilder, meta)
			targetBuilder.Build(extTarget)

			for _, p := range s.libProviders {
				s.libProvider(arch, extTarget, p, meta)
			}

			// Put dist under its own target
			distTarget := archTarget.Rule(arch.Target() + "_dist")
			if arch.IsWindows() {
				s.zip(arch, distTarget, meta)
			} else {
				s.tar(arch, distTarget, meta)
			}
		}
	}

	// BuildLocal then all should start with the build platform only
	if *s.BuildLocal {
		root.RemoveDependencies().
			AddDependency(s.buildArch.Target())
	}

	// Add any documentation
	docsBuilder := rootTarget.New()
	s.documentation.Do(docsBuilder, meta)

	root.Phony("docs")
	docsTarget := root.Rule("docs")
	docsBuilder.Build(docsTarget)

	if err := os.MkdirAll(filepath.Dir(*s.Dest), 0755); err != nil {
		return err
	}

	return os.WriteFile(*s.Dest, []byte(builder.Build()), 0644)
}

func (s *Build) allRule(arches []arch.Arch, builder makefile.Builder) (makefile.Builder, map[string]bool) {
	all := builder.Rule("all")
	platforms := make(map[string]bool)

	// Generate all target with either all or subset of platforms
	if *s.Platforms != "" {
		plats := strings.Split(*s.Platforms, " ")
		for _, arch := range arches {
			for _, plat := range plats {
				if strings.TrimSpace(plat) == arch.Platform() {
					all.AddDependency(arch.Target())
					platforms[arch.Target()] = true
				}
			}
		}
	}

	// If all is still empty then return it so the Operating System rules
	// will get added to it automatically
	if all.IsEmptyRule() {
		// In build local mode, add the local platform only to the build
		return all, platforms
	}

	// All is not empty so return the original builder
	return builder, platforms
}

func (s *Build) targetGroups(arches []arch.Arch, builder makefile.Builder, platforms map[string]bool) makefile.Map {
	osGroups := makefile.NewMap(builder)
	targetGroups := makefile.NewMap(builder)
	allPlatforms := len(platforms) == 0

	for _, arch := range arches {
		// Generate if all platforms or the platform exists
		if allPlatforms || platforms[arch.Target()] {

			goos := arch.GOOS
			if !osGroups.Contains(goos) {
				osGroups.Add(goos, func(builder makefile.Builder) makefile.Builder {
					return builder.Block().
						//Blank().
						//Comment("==================").
						//Comment(goos).
						//Comment("==================").
						Rule(goos, "init")
				})
			}

			target := arch.Target()
			if !targetGroups.Contains(target) {
				targetGroups.Add(target, func(_ makefile.Builder) makefile.Builder {
					return osGroups.Get(goos).
						Block().
						//Blank().
						//Comment("------------------").
						//Comment("%s %s", arch.GOOS, arch.Arch()).
						//Comment("------------------").
						Rule(target, "init")
				})
			}

		}
	}

	return targetGroups
}

func (s *Build) init(builder makefile.Builder) {
	builder.Rule("init").
		Mkdir(*s.Encoder.Dest, *s.Dist)
}

func (s *Build) callBuilder(builder makefile.Builder, action, cmd string, args ...string) {
	builder.Line("@$(BUILD) -d %s -%s %s %s", *s.Encoder.Dest, action, cmd, strings.Join(args, " "))
}

func (s *Build) clean(builder makefile.Builder) {
	rule := builder.Rule("clean").
		RM(s.cleanDirectories...)
	s.callBuilder(rule, "go", "clean", "--", "-testcache")
}

func (s *Build) test(builder makefile.Builder) {
	out := filepath.Join(*s.Encoder.Dest, "go-text.txt")

	rule := builder.Rule("test", "init").
		Mkdir(filepath.Dir(out))

	s.callBuilder(rule, "go", "test")
}

// Build a tool in go
func (s *Build) goBuild(arch arch.Arch, target makefile.Builder, tool string, _ *meta.Meta) {
	dest := arch.Tool(*s.Encoder.Dest, tool)

	rule := target.Rule(dest).
		Mkdir(filepath.Dir(dest))

	if arch.GOARM == "" {
		rule.Line("@$(BUILD) -d %s -go build %s %s %s", *s.Encoder.Dest, arch.GOOS, arch.GOARCH, tool)
	} else {
		rule.Line("@$(BUILD) -d %s -go build %s %s %s %s", *s.Encoder.Dest, arch.GOOS, arch.GOARCH, arch.GOARM, tool)
	}
}

// Add rules for a LibProvider
func (s *Build) libProvider(arch arch.Arch, target makefile.Builder, f LibProvider, _ *meta.Meta) {
	dest, args := f(arch.BaseDir(*s.Encoder.Dest))
	target.Rule(dest).
		Echo("GENERATE", strings.Join(strings.Split(dest, "/")[1:], " ")).
		Line("$(BUILD) -d %s %s", dest, strings.Join(args, " "))
}

// Add rule for a tar distribution
func (s *Build) tar(arch arch.Arch, target makefile.Builder, meta *meta.Meta) {
	archive := filepath.Join(
		*s.Dist,
		fmt.Sprintf("%s_%s_%s_%s%s.tgz", meta.PackageName, meta.Version, arch.GOOS, arch.GOARCH, arch.GOARM),
	)

	rule := target.Rule(archive)

	s.callBuilder(rule, "tar", archive, arch.BaseDir(*s.Encoder.Dest))
}

// Add rule for a zip distribution
func (s *Build) zip(arch arch.Arch, target makefile.Builder, meta *meta.Meta) {
	archive := filepath.Join(
		*s.Dist,
		fmt.Sprintf("%s_%s_%s_%s%s.zip", meta.PackageName, meta.Version, arch.GOOS, arch.GOARCH, arch.GOARM),
	)

	rule := target.Rule(archive)

	s.callBuilder(rule, "zip", archive, arch.BaseDir(*s.Encoder.Dest))
}

func (s *Build) platformIndex(arches []arch.Arch) error {
	var a []string
	a = append(a,
		"# Supported Platforms",
		"",
		"The following platforms are supported by virtue of how the build system works:",
		"",
		"| Operating System | CPU Architectures |",
		"| ---------------- | ----------------- |",
	)

	osCount := 0
	cpuCount := make(map[string]bool)

	larch := ""
	for _, arch := range arches {
		cpuCount[arch.Arch()] = true

		if arch.GOOS != larch {
			larch = arch.GOOS

			var as []string
			as = append(as, "|", larch, "|")
			osCount++
			for _, arch2 := range arches {
				if arch2.GOOS == larch {
					as = append(as, arch2.GOARCH+arch2.GOARM)
				}
			}
			as = append(as, "|")
			a = append(a, strings.Join(as, " "))
		}
	}

	a = append(a,
		"",
		fmt.Sprintf("Operating Systems %d CPU's %d", osCount, len(cpuCount)),
		"")

	return os.WriteFile("platforms.md", []byte(strings.Join(a, "\n")), 0644)
}

func (s *Build) jenkinsfile(arches []arch.Arch) error {

	builder := jenkinsfile.New()

	builder.Begin("properties([").
		Array().
		Begin("buildDiscarder(").
		Begin("logRotator(").
		Array().
		Property("artifactDaysToKeepStr", "").
		Property("artifactNumToKeepStr", "").
		Property("daysToKeepStr", "").
		Property("numToKeepStr", 10).
		End().End().
		Simple("disableConcurrentBuilds").
		Simple("disableResume").
		Begin("pipelineTriggers([").
		Simple("cron", `"H H * * *"`)

	node := builder.Node(*s.BuildNode)

	node.Stage("Checkout").
		Line("checkout scm")

	node.Stage("Init").
		Sh("make clean init")

	node.Stage("Test").
		Sh("make test")

	if *s.BuildLocal {
		// Build against the local platform only
		node.Stage("Build").
			Sh("make -f Makefile.gen all")
	} else {
		// Cross Build against the supported/requested platforms
		// Map of stages -> arch -> steps
		stages := make(map[string]*OsStage)

		for _, arch := range arches {
			if *s.Parallelize {
				// The older version, create a stage per OS then an inner
				// stage for each architecture which will be compiled in parallel
				stage := stages[arch.GOOS]
				if stage == nil {
					stage = NewOsStage(node, arch, arch.GOOS)
					stage.builder = stage.builder.Parallel()
				}
				stage1 := stage.add(node, arch, arch.Arch())
				if stage1 == nil {
					stage1 = &OsStage{
						arch:    arch,
						builder: stage.builder.Stage(arch.Arch()),
					}
				}
				stage1.builder.Sh("make -f Makefile.gen " + arch.Target())

				stages[arch.GOOS] = stage
			} else {
				// The new version unless overridden, run each OS/Arch as a
				// stage in sequence
				stage := stages[arch.Target()]
				if stage == nil {
					stage = NewOsStage(node, arch, arch.Target())
				}
				stage.builder.Sh("make -f Makefile.gen " + arch.Target())

				stages[arch.Target()] = stage
			}
		}

		// Sort stages so they are in sequence
		for _, s1 := range stages {
			s1.builder.Sort()
		}
	}

	// Add archiveArtifacts stage
	if *s.ArchiveArtifacts != "" {
		node.Stage("archiveArtifacts").
			ArchiveArtifacts(*s.ArchiveArtifacts)
	}

	return os.WriteFile("Jenkinsfile", []byte(builder.Build()), 0644)
}

func NewOsStage(node jenkinsfile.Builder, a arch.Arch, n string) *OsStage {
	return &OsStage{
		arch:    a,
		builder: node.Stage(n),
	}
}

type OsStage struct {
	arch     arch.Arch
	builder  jenkinsfile.Builder
	children map[string]*OsStage
}

func (s *OsStage) sort() {
	s.builder.Sort()
	if s.children != nil {
		for _, c := range s.children {
			c.sort()
		}
	}
}

func (s *OsStage) add(node jenkinsfile.Builder, a arch.Arch, n string) *OsStage {
	if s.children == nil {
		s.children = make(map[string]*OsStage)
	}
	if e, exists := s.children[n]; exists {
		return e
	}
	c := NewOsStage(node, a, n)
	s.children[n] = c
	return c
}
