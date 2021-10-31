package icloud

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

var Debug = false

func Marshal(v interface{}) []byte {
	if !Debug {
		data, err := json.Marshal(v)
		if err != nil {
			log.Fatalf("cannot marshal %v: %v", v, err)
		}
		return data
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("cannot marshal %v: %v", v, err)
	}
	data = append(data, '\n')
	return data
}
