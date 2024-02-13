package awesomemy

import (
	"time"

	"github.com/goccy/go-yaml"
)

// Duration represents a duration with YAML unmarshalling support.
type Duration struct {
	time.Duration
}

var _ yaml.BytesUnmarshaler = (*Duration)(nil)

// UnmarshalYAML unmarshals a YAML duration into a time.Duration.
func (d *Duration) UnmarshalYAML(b []byte) error {
	duration, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}

	d.Duration = duration

	return nil
}
