package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
)

const DefaultConfigPath = "config.store"

// just preventing that no one sees your actual password over your shoulder by accident
var configKey = []byte("xei6bi7miXieWoDathohHe1baeseifae")

type ConfigContainer struct {
	MinPagesLeft int
	Receiver     string

	SSHUser    string
	SSHKeyFile string

	PaperCutUsername string
	PaperCutPassword string

	TimeSlotMinutes int
}

func LoadOrCreateConfig(path string) *ConfigContainer {
	config, err := LoadConfig(path)
	if err != nil {
		config := CreateConfig()
		err := SaveConfig(path, config)
		if err != nil {
			fmt.Println("Failed to save config file:", err)
		}
	}
	return config
}

func CreateConfig() *ConfigContainer {
	fmt.Println("Could not load config file.")
	fmt.Println("A new config container is created.")
	fmt.Println("Your credentials will be encrypted when stored on disk.")

	fmt.Println()
	user := scanInput("PaperCut username: ")
	pass := scanPassword("PaperCut password: ")

	fmt.Println()
	sshUser := scanInput("Your TUD SSH username (leave empty to use PaperCut username): ")
	sshKeyFile := scanInput("Privatekey path (leave empty to use the keyagent): ")

	if sshUser == "" {
		sshUser = user
	}

	fmt.Println()
	receiver := scanInput("Who do you want so send your pages to: ")
	minPages := scanInt("Minimum amount of pages that should stay in your account: ")

	fmt.Println()
	fmt.Println("You can specify a time window ending at the turn of the month")
	fmt.Println("and papertrans will only transfer pages inside that time window.")
	minutes := scanIntOrDefault("Time window length in minutes (leave empty to transfer every time): ", 0)

	return &ConfigContainer {
		PaperCutUsername: user,
		PaperCutPassword: pass,
		SSHUser:          sshUser,
		SSHKeyFile:       sshKeyFile,
		MinPagesLeft:     minPages,
		Receiver:         receiver,
		TimeSlotMinutes:  minutes,
	}
}

func LoadConfig(path string) (*ConfigContainer, error) {
	config := new(ConfigContainer)
	err := loadEncrypted(path, config, configKey)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func SaveConfig(path string, config *ConfigContainer) error {
	return saveEncrypted(path, config, configKey)
}

func loadEncrypted(path string, object interface{}, key []byte) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	dec, err := decrypt(buf, key)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(dec)
	decoder := gob.NewDecoder(reader)
	err = decoder.Decode(object)
	return err
}

func saveEncrypted(path string, object interface{}, key []byte) error {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	encoder.Encode(object)

	enc, err := encrypt(buf.Bytes(), key)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, enc, 0700)
	return err
}
