package imgpack

import (
	"encoding/json"
	"os"
	"path"

	"github.com/VoileLab/goimgpack/internal/util"
)

type conf struct {
	// Scale is the scale factor of the application.
	Scale *float64 `json:"scale"`
}

func getConfPath() (string, error) {
	userConfDir, err := os.UserConfigDir()
	if err != nil {
		return "", util.Errorf("%w", err)
	}

	return path.Join(userConfDir, "goimgpack", "config.json"), nil
}

func getConf() (*conf, error) {
	confPath, err := getConfPath()
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	bs, err := os.ReadFile(confPath)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	var retConf conf
	if err := json.Unmarshal(bs, &retConf); err != nil {
		return nil, util.Errorf("%w", err)
	}

	return &retConf, nil
}

func setConf(c *conf) error {
	confPath, err := getConfPath()
	if err != nil {
		return util.Errorf("%w", err)
	}

	bs, err := json.Marshal(c)
	if err != nil {
		return util.Errorf("%w", err)
	}

	err = os.MkdirAll(path.Dir(confPath), 0755)
	if err != nil {
		return util.Errorf("%w", err)
	}

	if err := os.WriteFile(confPath, bs, 0644); err != nil {
		return util.Errorf("%w", err)
	}

	return nil
}
