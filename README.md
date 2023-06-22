# go-build

This is the build system for all of my go projects going forward.
Existing projects are going to be migrated to this system.

It supports the building of multiple executables, cross compiling against every
operating system and cpu architecture supported by the installed go environment.

At the time of writing with GO 1.20.5 that's 15 Operating Systems and 14 CPU architectures.

## Performing builds

In normal use, running `make clean all` will compile the project for every supported combination.

To build for just one then you need to tell make which platform(s) you want.

e.g.

    PLATFORMS=linux:amd64: make clean all

will compile just for Linux on the amd64 processor.

For multiple architectures then provide them on the same line.

e.g.

    ;PLATFORMS="linux:amd64: linux:arm64: darwin:arm64:" make clean all

will compile for Linux on the amd64 and arm64 processors as well as for MacOS on
arm64 (Apple Silicon).

The format for the platform declaration is operatingSystem:cpuArchitecture:variant

In most cases the variant is blank and is only used for 32-bit ARM processors.

e.g. `linux:arm:7` would be for the ARM7 processor used on later Raspberry PI's with a
32-bit operating system. Earlier PI's would need `linux:arm:6`.

NB:

* For Raspberry PI's with a 64-bit operating system you would use `linux:arm64:`
* For MacOS, use `darwin:amd64:` for Intel based Mac's and
  `darwin:arm64:` for Apple Silicon Mac's (tested with M1 and M1-Max)

## Layout of projects

There is only a single constraint on how a project is laid out:
Each binary to compile must be defined under the tools directory.

Specifically, for a tool called `mybinary` then you need to have it's `main()` method
defined in the file `tools/mybinary/bin/main.go`.

When the build environment runs it will look for all instances of main.go and generate
the appropriate entries in a Makefile for each tool.

Note: the `build` tool is special and will not be included in your final distributions.

# Configuring your project

You need to setup a few files within your project:

## Makefile

This is the bootstrap for the build environment.
You just need to copy the one in this project to the root of your project for it to work.

## .gitignore

You need to have the following entries in your `.gitignore` file, as these files are temporary.

    /build
    /builds/
    /dist/
    /Makefile.gen

* `build` is the build executable compiled and run by the environment,
* `builds` is the directory where your project will be built,
* `dist` is where tar and zip files of your project will be placed,
* `Makefile.gen` is the Makefile the environment creates to build your project.

## tools/build/bin/main.go

This is the entry point of the environment.
At a bare minimum for simple projects this should be a copy of what is here, specifically:

    package main
    
    import (
        "fmt"
        _ "github.com/peter-mount/go-build/tools/build"
        "github.com/peter-mount/go-kernel/v2"
        "os"
    )
    
    func main() {
        if err := kernel.Launch(); err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
    }

More complex projects can extend the environment to support deploying additional artifacts
in the build. This will be documented later but examples are [go-script](https://github.com/peter-mount/go-script)
and [piweather.center](https://github.com/peter-mount/piweather.center) which were the projects
the build environment was originally developed for.

# Generated files

In addition to the files included in `.gitignore` the environment generates two additional files
which should be committed with your project, as they only change when you change go version.

* `Jenkinsfile` is to allow [Jenkins](https://www.jenkins.io/) to build your project with the
  MultiBranchPipeline plugin.
* `platforms.md` is a markdown file listing each Operating System and CPU architecture the build
  environment will compile for when you do not declare which platform to use.

# Blocked platforms

There is a list of platforms supported by GO which are blocked by the environment.
These are listed in `util/arch/blocklist.go` and contain the following:

* `android:*:` and `ios:*:` - not sure how you would run anything on those platforms
* `js:*:` this is for web assembly - so unlikely to be useful for most projects
* `openbsd:mips64:` this is due to a bug in recent versions of go which can randomly
  [cause builds to fail](https://github.com/peter-mount/piweather.center/issues/1).
  This will be unblocked when they fix that issue. 