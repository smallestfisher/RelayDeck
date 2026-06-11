package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func LoadDotEnv() {
	for _, path := range dotEnvCandidates() {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			return
		}
	}
}

func dotEnvCandidates() []string {
	candidates := []string{}
	if cwd, err := os.Getwd(); err == nil {
		for dir := cwd; ; dir = filepath.Dir(dir) {
			candidates = append(candidates, filepath.Join(dir, ".env"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}
	_, file, _, ok := runtime.Caller(0)
	if ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
		candidates = append(candidates, filepath.Join(root, ".env"))
	}
	return candidates
}
