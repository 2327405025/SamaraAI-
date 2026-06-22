package image

import (
	"SamaraAI/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

const onnxRuntimeVersion = "1.22.0"

var (
	ortInitOnce sync.Once
	ortInitErr  error
)

func ensureOrtEnvironment() error {
	ortInitOnce.Do(func() {
		if path := resolveRuntimeLibPath(); path != "" {
			ort.SetSharedLibraryPath(path)
		}
		ortInitErr = ort.InitializeEnvironment()
		if ortInitErr != nil {
			ortInitErr = fmt.Errorf(
				"onnxruntime initialize error (need onnxruntime %s DLL/SO in project root or lib/, run scripts/download-models.cmd): %w",
				onnxRuntimeVersion,
				ortInitErr,
			)
		}
	})
	return ortInitErr
}

func resolveRuntimeLibPath() string {
	cfg := config.Get()
	if cfg != nil && cfg.Image.RuntimeLibPath != "" {
		return cfg.Image.RuntimeLibPath
	}

	names := []string{"onnxruntime.dll"}
	if runtime.GOOS != "windows" {
		names = []string{
			"libonnxruntime.so." + onnxRuntimeVersion,
			"libonnxruntime.so",
		}
	}

	candidates := make([]string, 0, len(names)*3)
	for _, name := range names {
		candidates = append(candidates, name, filepath.Join("lib", name))
	}
	if root := findModuleRoot(); root != "" {
		for _, name := range names {
			candidates = append(candidates, filepath.Join(root, name), filepath.Join(root, "lib", name))
		}
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, name := range names {
			candidates = append(candidates, filepath.Join(dir, name))
		}
	}

	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, err := filepath.Abs(c)
			if err != nil {
				return c
			}
			return abs
		}
	}
	return ""
}

func findModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
