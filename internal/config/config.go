package config

import (
    "encoding/json"
    "os"
)

type Secret struct {
    Path      string `json:"path"`
    Reference string `json:"reference"`
}

type Config struct {
    Secrets []Secret `json:"secrets"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}
