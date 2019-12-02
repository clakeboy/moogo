package common

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type SSHServer struct {
	Addr     string
	User     string
	Password string
	KeyPath  string
}
type SSHConfig struct {
	SSH   SSHServer
	Ports map[string]string
}

type SSHSession struct {
	client     *ssh.Client
	listenAddr string
	remoteAddr string
	status     int
	close      bool
}

const (
	SSHOpen = iota + 1
	SSHClosed
)

func NewSession(listen, remote string, client *ssh.Client) *SSHSession {
	return &SSHSession{
		client:     client,
		listenAddr: listen,
		remoteAddr: remote,
	}
}

func (s *SSHSession) handleConn(conn net.Conn) {
	log.Printf("accept %s", conn.RemoteAddr())
	remote, err := s.client.Dial("tcp", s.remoteAddr)
	if err != nil {
		log.Printf("dial %s error", s.remoteAddr)
		return
	}
	log.Printf("%s -> %s connected.", conn.RemoteAddr(), s.remoteAddr)
	wait := new(sync.WaitGroup)
	wait.Add(2)
	go func() {
		_, _ = io.Copy(remote, conn)
		_ = remote.Close()
		wait.Done()
	}()
	go func() {
		_, _ = io.Copy(conn, remote)
		_ = conn.Close()
		wait.Done()
	}()
	wait.Wait()
	log.Printf("%s -> %s closed", conn.RemoteAddr(), s.remoteAddr)
}

func (s *SSHSession) Run() error {
	l, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	fmt.Println("listen local ssh host:", s.listenAddr)
	s.status = SSHOpen
	for !s.close {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.handleConn(conn)
	}
	s.status = SSHClosed
	_ = l.Close()
	return nil
}

func (s *SSHSession) Close() {
	if s.client != nil {
		_ = s.client.Close()
	}
	s.close = true
}

func (s *SSHSession) CheckPort() (string, error) {
	addrs := strings.Split(s.listenAddr, ":")
	port, err := strconv.Atoi(addrs[1])
	if err != nil {
		return "", err
	}
	check := true
	for check {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			port++
			continue
		}
		_ = ln.Close()
		check = false
	}
	s.listenAddr = fmt.Sprintf("%s:%d", addrs[0], port)
	return s.listenAddr, nil
}

func LoginSSH(cfg *SSHServer) (*ssh.Client, error) {
	var methods []ssh.AuthMethod
	if cfg.KeyPath == "" && cfg.Password == "" {
		return nil, fmt.Errorf("empty private key and password")
	}

	if cfg.KeyPath != "" {
		key, err := ioutil.ReadFile(cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("unable to parse private key: %v", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if cfg.Password != "" {
		methods = append(methods, ssh.Password(cfg.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: methods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	return ssh.Dial("tcp", cfg.Addr, sshConfig)
}
