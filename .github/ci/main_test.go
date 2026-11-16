package main

import (
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMatrixRespectsPortsAndARMExceptions(t *testing.T) {
	config := configuration{ProjectName: "fileneko", Builds: []build{{
		GOOS:   []string{"linux", "windows", "openbsd"},
		GOARCH: []string{"amd64", "arm", "riscv64"},
		GOARM:  []string{"5", "6", "7"},
		Ignore: []goTarget{
			{GOOS: "windows", GOARCH: "arm"},
			{GOOS: "openbsd", GOARCH: "riscv64"},
			{GOOS: "linux", GOARCH: "arm", GOARM: "6"},
		},
	}}}
	ports := "linux/amd64 linux/arm linux/riscv64 windows/amd64 windows/arm openbsd/amd64 openbsd/riscv64"
	got, err := targets(config, ports)
	if err != nil {
		t.Fatal(err)
	}
	var archives []string
	for _, e := range got {
		archives = append(archives, e.Archive)
	}
	want := []string{
		"linux-amd64-fileneko.tar.gz",
		"linux-armv5-fileneko.tar.gz",
		"linux-armv7-fileneko.tar.gz",
		"linux-riscv64-fileneko.tar.gz",
		"openbsd-amd64-fileneko.tar.gz",
		"windows-amd64-fileneko.zip",
	}
	if !slices.Equal(archives, want) {
		t.Fatalf("archives = %v; want %v", archives, want)
	}
}

// Keep package verification sensitive to lost hardening flags, CPU baselines,
// source offers or redistribution notices when selecting an individual target.
func TestTargetSelectionPreservesReleaseSettings(t *testing.T) {
	source := []byte(`version: 2
project_name: fileneko
builds:
  - env: [CGO_ENABLED=0]
    binary: fileneko
    ldflags: ['-s -w -X version.SourceURL=source-archive']
    goos: [linux, windows]
    goarch: [amd64, arm]
    goarm: ['5', '6', '7']
    goamd64: [v1]
    goarm64: [v8.0]
archives:
  - files: [LICENSE, NOTICE, 'LICENSES/**']
`)
	selected, err := targetConfig(source, entry{goTarget: goTarget{GOOS: "linux", GOARCH: "arm", GOARM: "6"}})
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Builds []struct {
			Env     []string `yaml:"env"`
			Binary  string   `yaml:"binary"`
			LDFlags []string `yaml:"ldflags"`
			GOOS    []string `yaml:"goos"`
			GOARCH  []string `yaml:"goarch"`
			GOARM   []string `yaml:"goarm"`
			GOAMD64 []string `yaml:"goamd64"`
			GOARM64 []string `yaml:"goarm64"`
		} `yaml:"builds"`
		Archives []struct {
			Files []string `yaml:"files"`
		} `yaml:"archives"`
	}
	if err := yaml.Unmarshal(selected, &got); err != nil {
		t.Fatal(err)
	}
	b := got.Builds[0]
	for name, pair := range map[string][2][]string{
		"CGO":     {b.Env, {"CGO_ENABLED=0"}},
		"ldflags": {b.LDFlags, {"-s -w -X version.SourceURL=source-archive"}},
		"GOOS":    {b.GOOS, {"linux"}},
		"GOARCH":  {b.GOARCH, {"arm"}},
		"GOARM":   {b.GOARM, {"6"}},
		"AMD64":   {b.GOAMD64, {"v1"}},
		"ARM64":   {b.GOARM64, {"v8.0"}},
		"notices": {got.Archives[0].Files, {"LICENSE", "NOTICE", "LICENSES/**"}},
	} {
		if !slices.Equal(pair[0], pair[1]) {
			t.Errorf("%s = %v; want %v", name, pair[0], pair[1])
		}
	}
	if b.Binary != "fileneko" {
		t.Errorf("binary = %q", b.Binary)
	}
}
