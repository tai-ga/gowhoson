package whoson

import (
	katsubushi "github.com/kayac/go-katsubushi"
)

// NewIDGenerator is set id generator to IDGenerator.
func NewIDGenerator(workerID uint32) error {
	if IDGenerator == nil {
		idgen, err := katsubushi.NewGenerator(workerID)
		if err != nil {
			return err
		}
		IDGenerator = idgen
	}
	return nil
}
