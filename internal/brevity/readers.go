package brevity

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
)

func Read(specfile string, obj interface{}) error {
	switch filepath.Ext(specfile) {
	case ".json", ".jsn":
		return ReadJSON(specfile, obj)
	case ".yaml", ".yml":
		return ReadYAML(specfile, obj)
	case ".toml", ".tml":
		return ReadTOML(specfile, obj)
	default:
		return fmt.Errorf("unrecognized file format in file: %s", specfile)
	}
}

// ReadJSON spec from a file
func ReadJSON(specfile string, obj interface{}) error {
	file, err := os.Open(specfile)
	if err != nil {
		return fmt.Errorf("opening spec %s failed: %s", specfile, err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(obj); err != nil {
		return fmt.Errorf("decoding JSON spec %s failed: %s", specfile, err)
	}
	return nil
}

// ReadYAML spec from a file
func ReadYAML(specfile string, obj interface{}) error {
	file, err := os.Open(specfile)
	if err != nil {
		return fmt.Errorf("opening spec %s failed: %s", specfile, err)
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(obj); err != nil {
		return fmt.Errorf("decoding YAML spec %s failed: %s", specfile, err)
	}
	return nil
}

// ReadTOML spec from a file
func ReadTOML(specfile string, obj interface{}) error {
	file, err := os.Open(specfile)
	if err != nil {
		return fmt.Errorf("opening spec %s failed: %s", specfile, err)
	}
	defer file.Close()

	if err := toml.NewDecoder(file).Decode(obj); err != nil {
		return fmt.Errorf("decoding TOML spec %s failed: %s", specfile, err)
	}
	return nil
}
