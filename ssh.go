package papertrans

import (
	"context"
	"encoding/base64"
	"fmt"
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
		return errors.New("SSH host key didn't match")
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
	if file == "" {
		return nil
	}

	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Could not read privatekey:", err)
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		fmt.Println("Could not parse privatekey:", err)
		return nil
	}
	return ssh.PublicKeys(key)
}

func createSSHClient(host string, user string, keyFile string) (*ssh.Client, error) {
	var methods []ssh.AuthMethod

	if keyAgent := keyAgentAuth(); keyAgent != nil {
		methods = append(methods, keyAgent)
	}

	if pubKey := publicKeyAuth(keyFile); pubKey != nil {
		methods = append(methods, pubKey)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            methods,
		HostKeyCallback: checkHostKey,
	}

	return ssh.Dial("tcp", host, config)
}

func createSSHDialContext(client *ssh.Client) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return client.Dial(network, addr)
	}
}
