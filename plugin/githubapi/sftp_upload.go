package github

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"os"
)

func upload(bytes []byte, dir string, filename string) (os.FileInfo, error) {
	host := "192.168.0.113"
	port := 8022
	username := "u0_a111"
	password := "jing"
	filePath := fmt.Sprintf("%s/%s", dir, filename)

	// 创建 SSH 客户端连接。
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer conn.Close()

	// open an SFTP session over an existing ssh connection.
	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	err = client.MkdirAll(dir)
	if err != nil {
		return nil, err
	}

	// leave your mark
	f, err := client.Create(filePath)
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(bytes); err != nil {
		return nil, err
	}
	_ = f.Close()

	// check it's there
	fi, err := client.Lstat(filePath)
	if err != nil {
		return nil, err
	}
	return fi, nil
}
