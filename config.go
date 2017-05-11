package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
)

const credsContainerPath = "config.store"

// just preventing that no one sees your actual password over your shoulder by accident
var containerKey = []byte("xei6bi7miXieWoDathohHe1baeseifae")

type configContainer struct {
	MinPagesLeft int
	Receiver     string

	SSHUser    string
	SSHKeyFile string

	PaperCutUsername string
	PaperCutPassword string

	TimeSlotMinutes int
}

func getConfig() *configContainer {
	var container configContainer

	for loadEncrypted(credsContainerPath, &container, containerKey) != nil {
		createNewConfig()
	}

	return &container
}

func createNewConfig() {
	fmt.Println("Could not load config file.")
	fmt.Println("A new config container is created.")
	fmt.Println("Your credentials will be encrypted when stored on disk.")

	user := scanInput("PaperCut username: ")
	pass := scanPassword("PaperCut password: ")

	sshUser := scanInput("Your TUD SSH username: ")
	sshKeyFile := scanInput("Privatekey path (leave empty to use the keyagent): ")

	minPages := scanInt("Minimum amount of pages that should stay in your account: ")
	receiver := scanInput("To whom do you want so send your pages: ")

	fmt.Println("You can specify a time window ending at the turn of the month")
	fmt.Println("and papertrans will only transfer pages inside that time window.")
	minutes := scanIntOrDefault("Time window length in minutes (leave empty to transfer every time): ", 0)

	container := &configContainer{
		PaperCutUsername: user,
		PaperCutPassword: pass,
		SSHUser:          sshUser,
		SSHKeyFile:       sshKeyFile,
		MinPagesLeft:     minPages,
		Receiver:         receiver,
		TimeSlotMinutes:  minutes,
	}

	saveEncrypted(credsContainerPath, container, containerKey)
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
