package build

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/peter-mount/go-build/core"
	"github.com/peter-mount/go-kernel/v2/log"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Go struct {
	Encoder   *core.Encoder `kernel:"inject"`
	Go        *string       `kernel:"flag,go,call GO"`
	FailTests *bool         `kernel:"flag,go-test-fail,on test failure abort the build"`
}

func (s *Go) Start() error {
	if *s.Go != "" {
		return s.run()
	}
	return nil
}

func (s *Go) run() error {
	switch *s.Go {
	case "build":
		return s.build()

	case "clean":
		label("GO CLEAN", "-testcache")
		return s.cmd("go", "clean", "-testcache")

	case "test":
		return s.test()

	default:
		return fmt.Errorf("unknown GO command %q", *s.Go)
	}
}

func label(label, f string, a ...any) {
	fmt.Printf("%-8s ", label)
	fmt.Printf(f, a...)
	fmt.Println()
}

func (s *Go) cmd(name string, a ...string) error {
	cmd := exec.Command(name, a...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *Go) build() error {
	args := flag.Args()

	switch len(args) {
	case 3:
		return s.buildTool(args[0], args[1], "", args[2])

	case 4:
		return s.buildTool(args[0], args[1], args[2], args[3])

	default:
		return errors.New("-go build goos goarch [goarm] tool")
	}
}

func getEnv(k string) string {
	return strings.Trim(os.Getenv(k), `"`)
}

func (s *Go) buildTool(goos, goarch, goarm, tool string) error {
	src := filepath.Join("tools", tool, "bin/main.go")
	dst := filepath.Join(*s.Encoder.Dest, goos, goarch+goarm, "bin", tool)
	label("GO BUILD", dst)

	// Windows needs a file extension, legacy of MSDos and CP/M before that
	if goos == "windows" {
		dst = dst + ".exe"
	}

	// The os environment then add our vars
	env := append([]string{}, os.Environ()...)
	env = append(env, "CGO_ENABLED=0",
		"GOOS="+goos,
		"GOARCH="+goarch,
		"GOARM="+goarm,
	)

	var args []string
	args = append(args, "build")

	buildVersion := getEnv("BUILD_VERSION")
	buildTime := getEnv("BUILD_TIME")
	if buildVersion != "" && buildTime != "" {
		uid := getEnv("USER")
		if uid == "" {
			uuid := os.Getuid()
			if uuid >= 0 {
				uid = strconv.Itoa(uuid)
			} else {
				uid = "N/A"
			}
		}

		args = append(args,
			fmt.Sprintf(
				`-ldflags=-X '%s.Version=%s (%s %s %s %s %s)'`,
				getEnv("BUILD_PACKAGE_PREFIX"),
				tool,
				buildVersion,
				goos, goarch+goarm,
				uid,
				buildTime,
			))
	}

	args = append(args, "-o", dst, src)

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Env = env

	if log.IsVerbose() {
		log.Println(cmd.String())
	}

	return cmd.Run()
}

func (s *Go) test() error {
	var buf bytes.Buffer

	testOut := filepath.Join(*s.Encoder.Dest, "go-test.txt")
	f, err := os.Create(testOut)
	if err != nil {
		return err
	}
	defer f.Close()

	w := io.MultiWriter(&buf, f)
	cmd := exec.Command("go", "test", "./...")
	cmd.Stdout = w
	cmd.Stdin = os.Stdin
	cmd.Stderr = w

	if log.IsVerbose() {
		log.Println(cmd.String())
	}

	label("GO TEST", testOut)

	err = cmd.Run()
	if exit, ok := err.(*exec.ExitError); ok {
		fmt.Printf("Tests returned %d\n",
			exit.ExitCode())
		fmt.Println(buf.String())

		// Default don't fail the build on test failures
		if !*s.FailTests {
			return nil
		}
	}
	return err
}
