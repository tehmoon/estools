package esfilters

import (
	"io/ioutil"
	"encoding/json"
	"github.com/tehmoon/errors"
)

type Config struct {
	Filters *QueryFilters
}

func (c Config) ExportConfig() ([]byte, error) {
	var (
		filters json.RawMessage
		err error
	)

	filters, err = c.Filters.ExportConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error exporting filters")
	}

	raw := &ConfigRaw{
		Filters: &filters,
	}

	payload, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling config to JSON")
	}

	return payload, nil
}

func (c Config) ExportConfigToFile(p string) (error) {
	payload, err := c.ExportConfig()
	if err != nil {
		return errors.Wrap(err, "Error exporting config")
	}

	err = ioutil.WriteFile(p, payload, 0600)
	if err != nil {
		return errors.Wrap(err, "Error writing config to file")
	}

	return nil
}

type ConfigRaw struct {
	Filters *json.RawMessage `json:"filters"`
}

func ImportConfigFromFile(p string) (*Config, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading file")
	}

	return ImportConfig(data)
}

func ImportConfig(data []byte) (*Config, error) {
	config := NewConfig()

	raw := &ConfigRaw{}

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling data to JSON")
	}

	err = config.Filters.ImportConfig([]byte(*raw.Filters))
	if err != nil {
		return nil, errors.Wrap(err, "Error importing filters")
	}

	return config, nil
}

func NewConfig() (*Config) {
	return &Config{
		Filters: NewQueryFilters(),
	}
}
