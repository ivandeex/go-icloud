package icloud

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type dict map[string]interface{}

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

func (c *Client) getWebserviceURL(service string) (string, error) {
	return c.data.Webservices.URL(service)
}

func ReadLine(prompt string) string {
	var (
		line string
		err  error
	)
	for line == "" {
		fmt.Print(prompt)
		line, err = bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read line: %v", err)
		}
		line = strings.TrimSpace(line)
	}
	return line
}
