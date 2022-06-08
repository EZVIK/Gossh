# Gossh
SSH Agent 





```
import (
	"fmt"
	sshx "github.com/EZVIK/Gossh/sshx"
	"golang.org/x/crypto/ssh"
	"os"
	"time"
)


func main() {

	cfg := ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("password"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	gs := sshx.NewSSHClient("192.168.0.1", 22, &cfg)

	defer gs.Close()

	loginInfo, err := gs.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(loginInfo)

	resp, err := gs.Exec(sshx.CliCommands{
		Command: []string{
			"docker ps",
		},
		Timeout: time.Second * 60,
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	text := ""
	for _, r := range resp["docker ps"] {
		text += r + "\n"
	}

	fmt.Println(text)
}
```

