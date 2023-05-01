package github

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"testing"
)

func TestUpload(t *testing.T) {
	info, err := upload([]byte("Hello world!"), "qq/download", "hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info.Name())
}

func TestFtp(t *testing.T) {
	host := "192.168.0.113"
	port := 8022
	username := "u0_a111"
	password := "jing"
	//sourceFilePath := "/path/to/file/on/local/machine"
	//destinationFilePath := "/path/to/file/on/remote/server"

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
		t.Fatal(fmt.Errorf("failed to dial SSH: %w", err))
	}
	defer conn.Close()

	// open an SFTP session over an existing ssh connection.
	client, err := sftp.NewClient(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	////walk a directory
	//w := client.Walk("/home/qq")
	//for w.Step() {
	//	if w.Err() != nil {
	//		continue
	//	}
	//	t.Log(w.Path())
	//}

	// leave your mark
	f, err := client.Create("qq/hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("Hello world!")); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// check it's there
	fi, err := client.Lstat("qq/hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fi)
}
