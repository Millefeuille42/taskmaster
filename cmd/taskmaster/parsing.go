package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
)

type ProgramStatus struct {
	Running     bool
	ExitedEarly bool
	ExitCode    int
}

type Config struct {
	name        string
	pids        map[int]ProgramStatus
	Command     []string      `yaml:"command"`
	NumProcs    int           `yaml:"numprocs"`
	AutoStart   bool          `yaml:"autostart"`
	RestartWhen string        `yaml:"restart_when"`
	ExitCodes   []int         `yaml:"exit_codes"`
	StartTime   time.Duration `yaml:"start_time"`
	StopTime    time.Duration `yaml:"stop_time"`
	restart     int
	MaxRestarts int               `yaml:"max_restarts"`
	StopSignal  string            `yaml:"stop_signal"`
	Stdout      string            `yaml:"stdout"`
	Stderr      string            `yaml:"stderr"`
	Env         map[string]string `yaml:"env"`
	WorkDir     string            `yaml:"workdir"`
	Umask       int               `yaml:"umask"`
	lock        *sync.Mutex
}

func (c *Config) String() string {
	return fmt.Sprintf("%s", c.name)
}

func (c *Config) HasCriticalChange(oldConfig Config) bool {
	return !reflect.DeepEqual(c.Command, oldConfig.Command) ||
		c.NumProcs < oldConfig.NumProcs ||
		c.Stdout != oldConfig.Stdout ||
		c.Stderr != oldConfig.Stderr ||
		!reflect.DeepEqual(c.Env, oldConfig.Env) ||
		c.WorkDir != oldConfig.WorkDir ||
		c.Umask != oldConfig.Umask
}

func parseConfigFile(path string) (map[string]Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	var configs map[string]Config
	if err = yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in file %s: %v", path, err)
	}

	return configs, nil
}

func isYAMLFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".yaml" || ext == ".yml"
}

func parseDirectory(dir string) (map[string]Config, error) {
	configs := make(map[string]Config)

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %v", dir, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		if isYAMLFile(file.Name()) {
			fileConfigs, err := parseConfigFile(filePath)
			if err != nil {
				slog.Error(fmt.Sprintf("error processing file %s: %v", filePath, err))
				continue
			}
			for key, config := range fileConfigs {
				configs[key] = config
			}
		}
	}

	return configs, nil
}

func parseConfig() map[string]Config {
	allConfigs := make(map[string]Config)
	for _, arg := range flag.Args() {
		fileInfo, err := os.Stat(arg)
		if err != nil {
			slog.Error(fmt.Sprintf("error accessing %s: %v", arg, err))
			continue
		}

		var configs map[string]Config
		if fileInfo.IsDir() {
			configs, err = parseDirectory(arg)
		} else if isYAMLFile(arg) {
			configs, err = parseConfigFile(arg)
		} else {
			slog.Warn(fmt.Sprintf("skipping non-YAML file %s", arg))
			continue
		}

		if err != nil {
			slog.Error(fmt.Sprintf("error processing argument %s: %v", arg, err))
			continue
		}
		for key, config := range configs {
			config.name = key
			config.pids = make(map[int]ProgramStatus)
			config.lock = new(sync.Mutex)
			allConfigs[key] = config
		}
	}

	return allConfigs
}
