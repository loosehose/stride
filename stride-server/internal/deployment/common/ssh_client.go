// Common SSH functionality, such as establishing an SSH connection

/*
This file would contain common functionality required to establish SSH connections to your agents.
This might include functions to execute commands over SSH, transfer files, etc.
*/

package common

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHClient struct to hold connection details and the ssh.Client
type SSHClient struct {
	Host       string
	Port       string
	User       string
	PrivateKey string
	client     *ssh.Client
}

func addHostKey(host string) error {
	cmd := exec.Command("ssh-keyscan", "-H", host, ">>", "~/.ssh/known_hosts")
	return cmd.Run()
}

// NewSSHClient creates and returns a new SSH client after establishing a connection
func NewSSHClient(host, port, user, privateKeyPath string) (*SSHClient, error) {

	if err := addHostKey(host); err != nil {
		return nil, fmt.Errorf("failed to add host key: %v", err)
	}
	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: Use a more secure method in production.
	}

	sshClient, err := ssh.Dial("tcp", net.JoinHostPort(host, port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return &SSHClient{
		Host:       host,
		Port:       port,
		User:       user,
		PrivateKey: string(privateKey),
		client:     sshClient,
	}, nil
}

// connect establishes an SSH connection and returns an SSH session
func (c *SSHClient) connect() (*ssh.Session, error) {
	signer, err := ssh.ParsePrivateKey([]byte(c.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", net.JoinHostPort(c.Host, c.Port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	// Store the raw *ssh.Client in the struct for later use
	c.client = sshClient

	// Create a new session from the established SSH client
	session, err := sshClient.NewSession()
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return session, nil
}

// ExecuteCommand executes a command on the remote SSH server and returns its output
func (c *SSHClient) ExecuteCommand(cmd string) (string, error) {
    session, err := c.connect()
    if err != nil {
        return "", err
    }
    defer session.Close()

    var stdout, stderr bytes.Buffer
    session.Stdout = &stdout
    session.Stderr = &stderr

    if err := session.Run(cmd); err != nil {
        return "", fmt.Errorf("command '%s' failed: %v\nstdout: %s\nstderr: %s", cmd, err, stdout.String(), stderr.String())
    }

    return stdout.String(), nil
}

// TransferFile transfers a file to the remote SSH server using SFTP.
func (c *SSHClient) TransferFile(localPath, remotePath string) error {
	// Ensure there's an active SSH client connection
	if c.client == nil {
		return fmt.Errorf("SSH client not connected")
	}

	// Create a new SFTP client from the existing SSH client
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	// Open the source (local) file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	// Create the destination (remote) file
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	// Copy the local file to the remote location
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	return nil
}

// LoadFile reads a file from the given path and returns its content as a string.
func LoadFile(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file at %s: %v", filePath, err)
	}
	return string(content), nil
}
