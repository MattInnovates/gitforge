package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/crypto/ssh"

	"github.com/MattInnovates/gitforge/core/config"
	"github.com/MattInnovates/gitforge/core/storage/database"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.OpenAndMigrate(context.Background(), cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	repoRoot, err := filepath.Abs(cfg.ReposPath)
	if err != nil {
		log.Fatalf("resolve repos path: %v", err)
	}
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		log.Fatalf("create repos path: %v", err)
	}

	hostSigner, err := generateHostSigner()
	if err != nil {
		log.Fatalf("generate host signer: %v", err)
	}

	sshConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			fingerprint := ssh.FingerprintSHA256(key)
			log.Printf("ssh auth accepted user=%s key=%s", conn.User(), fingerprint)
			return &ssh.Permissions{Extensions: map[string]string{"key_fingerprint": fingerprint}}, nil
		},
	}
	sshConfig.AddHostKey(hostSigner)

	addr := fmt.Sprintf(":%d", cfg.GitSSHPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen on %s: %v", addr, err)
	}
	defer listener.Close()

	log.Printf("git-ssh listening on %s (repos=%s)", addr, repoRoot)
	for {
		netConn, err := listener.Accept()
		if err != nil {
			log.Printf("accept connection: %v", err)
			continue
		}

		go handleSSHConn(netConn, sshConfig, repoRoot)
	}
}

func handleSSHConn(netConn net.Conn, sshConfig *ssh.ServerConfig, repoRoot string) {
	defer netConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, sshConfig)
	if err != nil {
		log.Printf("ssh handshake failed: %v", err)
		return
	}
	defer sshConn.Close()

	log.Printf("ssh connection established from %s as %s", sshConn.RemoteAddr(), sshConn.User())
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "only session channels are supported")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("accept channel: %v", err)
			continue
		}

		go handleSession(channel, requests, repoRoot)
	}
}

func handleSession(channel ssh.Channel, requests <-chan *ssh.Request, repoRoot string) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			var payload struct {
				Command string
			}
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				_ = req.Reply(false, nil)
				writeSSHStderr(channel, "invalid exec payload\n")
				sendExitStatus(channel, 1)
				return
			}

			_ = req.Reply(true, nil)
			exitCode := runGitExec(channel, payload.Command, repoRoot)
			sendExitStatus(channel, exitCode)
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func runGitExec(channel ssh.Channel, rawCommand, repoRoot string) uint32 {
	service, repo, err := parseGitCommand(rawCommand)
	if err != nil {
		writeSSHStderr(channel, "invalid command: "+err.Error()+"\n")
		return 1
	}

	repoPath, err := resolveRepoPath(repoRoot, repo)
	if err != nil {
		writeSSHStderr(channel, "invalid repository path: "+err.Error()+"\n")
		return 1
	}

	subcommand := allowedGitServices[service]
	cmd := exec.Command("git", subcommand, repoPath)
	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel.Stderr()

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return uint32(exitErr.ExitCode())
		}

		writeSSHStderr(channel, "git command failed: "+err.Error()+"\n")
		return 1
	}

	return 0
}

func sendExitStatus(channel ssh.Channel, code uint32) {
	_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct {
		Status uint32
	}{Status: code}))
}

func writeSSHStderr(channel ssh.Channel, message string) {
	_, _ = io.WriteString(channel.Stderr(), message)
}
