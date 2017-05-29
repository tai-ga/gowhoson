package whoson

import (
	katsubushi "github.com/kayac/go-katsubushi"
)

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
