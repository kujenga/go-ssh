package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// based off of: http://blog.ralch.com/tutorial/golang-ssh-connection/

func main() {
	fmt.Println("get ready to ssh!")

	sshConfig := &ssh.ClientConfig{
		User: os.Getenv("USER"),
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
	}

	fmt.Print("Enter hostname: ")
	var host string
	if _, err := fmt.Scanln(&host); err != nil {
		log.Fatalln("error reading in hostname:", err)
	}

	connection, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		log.Fatalln("error dialing:", err)
	}
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		log.Fatalln("error making new session:", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		log.Fatalln("request for pseudo terminal failed:", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatalln("Unable to setup stdin for session:", err)
	}
	go io.Copy(stdin, os.Stdin)

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatalln("Unable to setup stdout for session:", err)
	}
	go io.Copy(os.Stdout, stdout)

	stderr, err := session.StderrPipe()
	if err != nil {
		log.Fatalln("Unable to setup stderr for session:", err)
	}
	go io.Copy(os.Stderr, stderr)

	err = session.Run("ls -l $LC_USR_DIR")
	if err != nil {
		log.Fatalln("error running command:", err)
	}
}

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}
