package main

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"net"
	"os"

	"errors"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func checkHostKey(dialAddr string, addr net.Addr, key ssh.PublicKey) error {
	baseKey := base64.StdEncoding.EncodeToString(key.Marshal())
	if baseKey != sshHostKey {
		return errors.New("ssh host key didn't match")
	}
	return nil
}

func keyAgentAuth() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func publicKeyAuth(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func createSSHClient(host string, user string, keyFile string) *ssh.Client {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			keyAgentAuth(),
			publicKeyAuth(keyFile),
		},
		HostKeyCallback: checkHostKey,
	}

	conn, err := ssh.Dial("tcp", host, config)
	check(err, "Failed to connect to SSH server")
	return conn
}

func createSSHDialContext(client *ssh.Client) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return client.Dial(network, addr)
	}
}
