package extensions

import (
	"os"
	"path/filepath"
	"sort"
)

// CollectAll walks base + every profile + every flavor under sourceDir, loads
// each extensions.yaml and returns the de-duplicated union of all Plugins and
// MCP servers (first-seen-name wins, matching Merge semantics).
//
// Unlike deploy.installExtensions — which only loads the active profile, its
// parents and the selected flavors — this loads EVERY profile and EVERY flavor.
// base is loaded first so its canonical definitions win on name conflicts.
func CollectAll(sourceDir string) (Extensions, error) {
	paths := []string{filepath.Join(sourceDir, "base", "extensions.yaml")}

	for _, group := range []string{"profiles", "flavors"} {
		names, err := subdirs(filepath.Join(sourceDir, group))
		if err != nil {
			return Extensions{}, err
		}
		for _, name := range names {
			paths = append(paths, filepath.Join(sourceDir, group, name, "extensions.yaml"))
		}
	}

	var layers []Extensions
	for _, path := range paths {
		ext, err := Load(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return Extensions{}, err
		}
		layers = append(layers, ext)
	}

	return Merge(layers), nil
}

// subdirs returns the sorted directory names directly under dir, or nil if dir
// does not exist.
func subdirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
