package SSH_SIM

import "strconv"

// Node 节点
type Node struct {
	ID       string // id
	TaskId   string
	DeviceId string
	Host     string // IP:PORT 127.0.0.1:22
	Port     uint16 // 0-65535 // port range 1024 to 49151.
	Command  []byte // docker run -exec ss
	Result   []string
}

func (n *Node) GetHostAndPort() string {
	return n.Host + ":" + strconv.Itoa(int(n.Port))
}
