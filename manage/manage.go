package manage

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Manager struct {
	Conn            *net.UDPConn
	receiveChan     chan string
	receiveDataChan chan string
	RemoteAddr      *net.UDPAddr
	LocalAddr       *net.UDPAddr
}

func NewManager(host string) (*Manager, error) {
	m := Manager{}
	var err error
	m.LocalAddr, err = net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	m.Conn, err = net.ListenUDP("udp", m.LocalAddr)
	if err != nil {
		return nil, err
	}

	m.RemoteAddr, err = net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, err
	}

	m.receiveChan = make(chan string)

	return &m, nil
}

func (m *Manager) Receive() (string, error) {
	dst := make([]byte, 2000)
	_, _, err := m.Conn.ReadFromUDP(dst)
	if err != nil {
		fmt.Println("Error receiving: ", err)
		return "", err
	}
	dst = bytes.TrimRight(dst, "\x00")
	return string(dst), nil
}

func (m *Manager) RequestRaw(data string) error {
	_, err := m.Conn.WriteToUDP([]byte(data), m.RemoteAddr)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) RequestReceive(data string) (string, error) {
	err := m.RequestRaw(data)
	if err != nil {
		return "", err
	}
	return m.Receive()
}

func (m *Manager) ReliableRequestReceive(data string) (string, error) {
	m.receiveDataChan = make(chan string, 1)
	go func() {
		reply, _ := m.Receive()
		select {
		case m.receiveDataChan <- reply:
		default:
			return
		}
	}()
	ctr := 0
	for {
		err := m.RequestRaw(data)
		if err != nil {
			return "", nil
		}
		timeout := time.After(500 * time.Millisecond)
		select {
		case reply := <-m.receiveDataChan:
			close(m.receiveDataChan)
			return reply, nil
		case <-timeout:
			ctr += 1
			if ctr == 5 {
				close(m.receiveDataChan)
				return "", fmt.Errorf("timeout")
			}
		}
	}
}

func (m *Manager) Auth() (string, error) {
	response, err := m.ReliableRequestReceive("HF-A11ASSISTHREAD")
	if err != nil {
		return "", err
	}
	// parts: ip, mac, software version
	parts := strings.Split(response, ",")
	if len(parts) < 3 {
		return "", fmt.Errorf("Unparsable response from lightbulb: %s", response)
	}
	if parts[2] != "HF-LPB100-ZJ200" {
		return "", fmt.Errorf("Unknown firmware version: %s", parts[2])
	}
	response, err = m.ReliableRequestReceive("+ok")
	if err != nil {
		return "", err
	}
	if response != "+ERR=-1\n\n" {
		return "", fmt.Errorf("Unexpected response to +ok: %x", response)
	}
	return parts[1], nil
}

func (m *Manager) Help() (string, error) {
	_, err := m.Auth()
	if err != nil {
		return "", err
	}
	lines := ""
	err = m.RequestRaw("AT+H\r")
	if err != nil {
		return "", err
	}
	for {
		line, err := m.Receive()
		if err != nil {
			return "", err
		}
		if line == "+ok\r\n\r\n\r\n" {
			lines = strings.TrimRight(lines, "\r\n")
			return lines, nil
		} else {
			lines += line
		}
	}
	return lines, nil
}

func (m *Manager) GetWSInfo() (string, string, error) {
	_, err := m.Auth()
	if err != nil {
		return "", "", err
	}
	ssid, err := m.ReliableRequestReceive("AT+WSSSID\n")
	if err != nil {
		return "", "", err
	}
	password, err := m.ReliableRequestReceive("AT+WSKEY\n")
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(ssid), strings.TrimSpace(password), nil
}

func (m *Manager) HTTPSend(host, port, method, path, conn, userAgent, data string) (string, error) {
	_, err := m.Auth()
	if err != nil {
		return "", err
	}
	r, err := m.ReliableRequestReceive("AT+HTTPURL=" + host + "," + port + "\n")
	if err != nil {
		return "", err
	}
	fmt.Println("url")
	fmt.Println(r)
	r, err = m.ReliableRequestReceive("AT+HTTPTP=" + method + "\n")
	if err != nil {
		return "", err
	}
	fmt.Println("tp")
	fmt.Println(r)
	r, err = m.ReliableRequestReceive("AT+HTTPPH=" + path + "\n")
	if err != nil {
		return "", err
	}
	fmt.Println("ph")
	fmt.Println(r)
	r, err = m.ReliableRequestReceive("AT+HTTPCN=" + conn + "\n")
	if err != nil {
		return "", err
	}
	fmt.Println("cn")
	fmt.Println(r)
	r, err = m.ReliableRequestReceive("AT+HTTPUA=" + userAgent + "\n")
	if err != nil {
		return "", err
	}
	if data != "" {
		r, err = m.RequestReceive("AT+HTTPDT=" + data + "\n")
		fmt.Println("dt")
		fmt.Println(r)
		if err != nil {
			return "", err
		}
	} else {
		r, err = m.RequestReceive("AT+HTTPDT\n")
		fmt.Println("dt")
		fmt.Println(r)
		if err != nil {
			return "", err
		}
	}

	go m.ShellReceiver()
	response := <-m.receiveChan
	return response, nil
}

func (m *Manager) doReceive() {
	for {
		data, _ := m.Receive()
		select {
		case m.receiveDataChan <- data:
		default:
			return
		}
	}
}

func (m *Manager) ShellReceiver() {
	m.receiveDataChan = make(chan string, 100)
	go m.doReceive()
	accum := ""
	for {
		timeout := time.After(time.Second)
		select {
		case data := <-m.receiveDataChan:
			accum += data
		case <-timeout:
			if accum != "" {
				m.receiveChan <- accum
				accum = ""
				close(m.receiveDataChan)
				m.receiveDataChan = make(chan string, 100)
				go m.doReceive()
			}
		}
	}
}

func (m *Manager) Shell() error {
	mac, err := m.Auth()
	if err != nil {
		return err
	}
	fmt.Printf("Connected to %s\n", mac)

	go m.ShellReceiver()

	consolereader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := consolereader.ReadString('\n')
		if err != nil {
			return err
		}
		input = input + "\r"
		fmt.Println(input)
		err = m.RequestRaw(input)
		if err != nil {
			return err
		}
		fmt.Println(strings.TrimSpace(<-m.receiveChan))
	}
}
