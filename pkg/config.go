package main

import (
  "encoding/json"
  "os"
)

type Config struct {
  Server             string            `json:"Server"`
  Port               string            `json:"Port"`
  Creds              map[string]string `json:"Creds"`
}

func LoadConfig(filename string) (*Config, error) {
  configFile, err := os.Open(filename)
  defer configFile.Close()
  if err != nil {
    return nil, err
  }

  config := Config{}
  jsonParser := json.NewDecoder(configFile)
  jsonParser.Decode(&config)

  return &config, nil
}

