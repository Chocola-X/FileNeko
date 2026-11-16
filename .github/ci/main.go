// CI derives its matrix and per-target configuration from the release config.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type goTarget struct {
	GOOS   string `yaml:"goos"`
	GOARCH string `yaml:"goarch"`
	GOARM  string `yaml:"goarm"`
}

type build struct {
	GOOS   []string   `yaml:"goos"`
	GOARCH []string   `yaml:"goarch"`
	GOARM  []string   `yaml:"goarm"`
	Ignore []goTarget `yaml:"ignore"`
}

type configuration struct {
	ProjectName string  `yaml:"project_name"`
	Builds      []build `yaml:"builds"`
}

type entry struct {
	Target  string `json:"target"`
	Archive string `json:"archive"`
	goTarget
}

func targets(config configuration, supported string) ([]entry, error) {
	if len(config.Builds) != 1 || config.ProjectName == "" {
		return nil, fmt.Errorf("CI expects one Go build and a project_name")
	}
	b := config.Builds[0]
	ports := strings.Fields(supported)
	var result []entry
	for _, goos := range b.GOOS {
		for _, arch := range b.GOARCH {
			if !slices.Contains(ports, goos+"/"+arch) {
				continue
			}
			arms := []string{""}
			if arch == "arm" {
				arms = b.GOARM
			}
			for _, arm := range arms {
				target := goTarget{goos, arch, arm}
				if slices.ContainsFunc(b.Ignore, func(ignore goTarget) bool {
					return (ignore.GOOS == "" || ignore.GOOS == goos) &&
						(ignore.GOARCH == "" || ignore.GOARCH == arch) &&
						(ignore.GOARM == "" || ignore.GOARM == arm)
				}) {
					continue
				}
				id := goos + "-" + arch
				if arm != "" {
					id += "v" + arm
				}
				ext := ".tar.gz"
				if goos == "windows" {
					ext = ".zip"
				}
				result = append(result, entry{id, id + "-" + config.ProjectName + ext, target})
			}
		}
	}
	slices.SortFunc(result, func(a, b entry) int { return strings.Compare(a.Target, b.Target) })
	if len(result) == 0 {
		return nil, fmt.Errorf("release config produced an empty build matrix")
	}
	return result, nil
}

func targetConfig(source []byte, selected entry) ([]byte, error) {
	var config map[string]interface{}
	if err := yaml.Unmarshal(source, &config); err != nil {
		return nil, err
	}
	builds, ok := config["builds"].([]interface{})
	if !ok || len(builds) != 1 {
		return nil, fmt.Errorf("CI expects exactly one build")
	}
	b, ok := builds[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Go build configuration")
	}
	b["goos"] = []string{selected.GOOS}
	b["goarch"] = []string{selected.GOARCH}
	if selected.GOARM != "" {
		b["goarm"] = []string{selected.GOARM}
	}
	// All other settings, including ldflags, CPU baselines and archive contents,
	// stay inherited from the same configuration used for releases.
	return yaml.Marshal(config)
}

func run(args []string) error {
	if len(args) != 1 && len(args) != 3 {
		return fmt.Errorf("usage: go run ./.github/ci matrix | config TARGET OUTPUT")
	}
	source, err := os.ReadFile(".goreleaser.yml")
	if err != nil {
		return err
	}
	var config configuration
	if err := yaml.Unmarshal(source, &config); err != nil {
		return err
	}
	ports, err := exec.Command("go", "tool", "dist", "list").Output()
	if err != nil {
		return err
	}
	matrix, err := targets(config, string(ports))
	if err != nil {
		return err
	}
	if args[0] == "matrix" && len(args) == 1 {
		// Only public fields belong in the Actions matrix.
		entries := make([]map[string]string, 0, len(matrix))
		for _, e := range matrix {
			entries = append(entries, map[string]string{"target": e.Target, "archive": e.Archive})
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"include": entries})
	}
	if args[0] == "config" && len(args) == 3 {
		for _, e := range matrix {
			if e.Target == args[1] {
				data, err := targetConfig(source, e)
				if err != nil {
					return err
				}
				return os.WriteFile(args[2], data, 0o600)
			}
		}
		return fmt.Errorf("target %q is not enabled in the release configuration", args[1])
	}
	return fmt.Errorf("unknown CI command %q", args[0])
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
