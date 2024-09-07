package network

import (
	"bytes"
	"edigo/pkg/crdt"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	broadcastAddr = ""
	broadcastMin  = 12340
	broadcastMax  = 12399 // Increased range
	tcpBasePort   = 12346
	maxAttempts   = 20 // Increased max attempts
)

type Network struct {
	IsHost         bool
	ID             string
	Host           net.Conn           // isHost = false
	Clients        []net.Conn         // isHost = true
	Sessions       map[string]Session // found connections
	NewConnection  chan net.Conn
	CurrentSession string // "" -> keine Session offen
	UdpPort        int
}

type Session struct {
	Name string
	IP   string
	Port int
}

var (
	sessionMutex sync.Mutex
	Connections  []net.Conn // active connections
	Quit         = make(chan bool)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRandomPort(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func NewNetwork() *Network {
	broadcastAddr, err := getBroadcastAddress()
	if err != nil {
		fmt.Printf("Fehler beim Ermitteln der Broadcast-Adresse: %v\n", err)
		os.Exit(1)
	}

	udpPort := 0

	for attempt := 0; attempt < maxAttempts; attempt++ {
		port := getRandomPort(broadcastMin, broadcastMax)
		addr, err := net.ResolveUDPAddr("udp", broadcastAddr+":"+strconv.Itoa(port))
		if err != nil {
			fmt.Printf("Fehler beim Auflösen der Broadcast-Adresse für Port %d: %v\n", port, err)
			continue
		}

		conn, err := net.ListenUDP("udp", addr)
		if err == nil {
			udpPort = port
			conn.Close()
			break
		}
		fmt.Printf("Konnte nicht auf Port %d lauschen: %v\n", port, err)
	}

	if udpPort == 0 {
		fmt.Printf("Kein freier Port im Bereich %d-%d verfügbar nach %d Versuchen.\n", broadcastMin, broadcastMax, maxAttempts)
		os.Exit(1)
	}

	network := &Network{IsHost: false, ID: generateNetworkID(), Sessions: make(map[string]Session), UdpPort: udpPort, CurrentSession: "", NewConnection: make(chan net.Conn)}
	return network
}

func (network *Network) ListenForBroadcasts() {
	addr, err := net.ResolveUDPAddr("udp", broadcastAddr+":"+strconv.Itoa(network.UdpPort))
	if err != nil {
		fmt.Printf("Fehler beim Auflösen der Broadcast-Adresse: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Fehler beim Öffnen der UDP-Verbindung: %v\n", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Fehler beim Lesen der UDP-Nachricht: %v\n", err)
			continue
		}

		isLocalAddr, err := isLocalAddress(remoteAddr.IP.String())
		if err != nil {
			fmt.Printf("Fehler beim Überprüfen der lokalen Adresse: %v\n", err)
			continue
		}

		message := string(buffer[:n])
		parts := strings.Split(message, "|")

		if len(parts) != 4 {
			fmt.Printf("Ungültiges Nachrichtenformat empfangen: %s\n", message)
			continue
		}

		udpPort, err := strconv.Atoi(parts[3])
		if err != nil {
			fmt.Printf("Fehler beim Konvertieren des UDP-Ports: %v\n", err)
			continue
		}

		if isLocalAddr && udpPort == network.UdpPort {
			continue
		}

		if parts[0] == "SESSION" {
			sessionName := parts[1]
			port, err := strconv.Atoi(parts[2])
			if err != nil {
				fmt.Printf("Fehler beim Konvertieren des Ports: %v\n", err)
				continue
			}
			sessionMutex.Lock()
			network.Sessions[sessionName] = Session{Name: sessionName, IP: remoteAddr.IP.String(), Port: port}
			sessionMutex.Unlock()
		}
	}
}

func (network *Network) BroadcastSession(rga *crdt.RGA) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Printf("Fehler beim Starten des TCP-Listeners: %v\n", err)
		return
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	sessionName := fmt.Sprintf("Session-%d", port)

	go func() {
		for {
			select {
			case <-Quit:
				return
			default:
				for udpPort := broadcastMin; udpPort <= broadcastMax; udpPort++ {
					message := []byte(fmt.Sprintf("SESSION|%s|%d|%d", sessionName, port, network.UdpPort))
					addr, err := net.ResolveUDPAddr("udp", broadcastAddr+":"+strconv.Itoa(udpPort))
					if err != nil {
						fmt.Printf("Fehler beim Auflösen der Broadcast-Adresse für Port %d: %v\n", udpPort, err)
						continue
					}

					conn, err := net.DialUDP("udp", nil, addr)
					if err != nil {
						fmt.Printf("Fehler beim Öffnen der UDP-Verbindung für Port %d: %v\n", udpPort, err)
						continue
					}
					_, err = conn.Write(message)
					if err != nil {
						fmt.Printf("Fehler beim Senden der Broadcast-Nachricht an Port %d: %v\n", udpPort, err)
					}
				}
				time.Sleep(3 * time.Second)
			}
		}
	}()

	for {
		select {
		case <-Quit:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Fehler beim Akzeptieren der Verbindung: %v\n", err)
				continue
			}
			network.Clients = append(network.Clients, conn)
			network.NewConnection <- conn
			network.IsHost = true
			SendInitRGA(*rga, conn)
		}
	}
}

func (network *Network) JoinSession(sessionName string) crdt.RGA {
	sessionMutex.Lock()
	session, exists := network.Sessions[sessionName]
	sessionMutex.Unlock()

	if !exists {
		fmt.Println("Sitzung nicht gefunden.")
		return crdt.RGA{}
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", session.IP, session.Port))
	if err != nil {
		fmt.Printf("Fehler beim Verbinden mit der Sitzung: %v\n", err)
		return crdt.RGA{}
	}

	// First read the size of the incoming data
	var size int64
	err = binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		conn.Close()
		return crdt.RGA{}
	}

	// Now read the exact amount of data
	tmp := make([]byte, size)
	_, err = io.ReadFull(conn, tmp)
	if err != nil {
		fmt.Printf("Fehler beim Lesen der Initialdaten: %v\n", err)
		conn.Close()
		return crdt.RGA{}
	}

	tmpbuff := bytes.NewBuffer(tmp)
	tmpstruct := new(crdt.RGA)
	gobobj := gob.NewDecoder(tmpbuff)

	err = gobobj.Decode(tmpstruct)
	if err != nil {
		fmt.Printf("Fehler beim Dekodieren der RGA-Daten: %v\n", err)
		conn.Close()
		return crdt.RGA{}
	}

	network.Host = conn
	network.NewConnection <- conn
	network.IsHost = false
	network.CurrentSession = session.Name

	// Verify the integrity of the received RGA
	if !tmpstruct.VerifyIntegrity() {
		return crdt.RGA{}
	}

	return *tmpstruct
}

func (network *Network) SendOperation(op crdt.Operation, conn net.Conn) {
	bin_buf := new(bytes.Buffer)
	gobobj := gob.NewEncoder(bin_buf)
	err := gobobj.Encode(op)
	if err != nil {
		fmt.Printf("Fehler beim Kodieren der Operation: %v\n", err)
		return
	}
	_, err = conn.Write(bin_buf.Bytes())
	if err != nil {
		fmt.Printf("Fehler beim Senden der Operation: %v\n", err)
	}
}

func (network *Network) CloseAsHost() {
	for _, conn := range network.Clients {
		conn.Close()
	}
	network.CurrentSession = ""
	network.IsHost = false
}

func (network *Network) HostClosedSession() {
	if network.Host != nil {
		network.Host.Close()
	}
	network.Host = nil
	network.CurrentSession = ""
}

func (network *Network) RemoveClient(conn net.Conn) {
	for i, c := range network.Clients {
		if c == conn {
			network.Clients = append((network.Clients)[:i], (network.Clients)[i+1:]...)
			conn.Close()
			return
		}
	}
}

func SendInitRGA(rga crdt.RGA, conn net.Conn) {
	bin_buf := new(bytes.Buffer)
	gobobj := gob.NewEncoder(bin_buf)
	err := gobobj.Encode(rga)
	if err != nil {
		fmt.Printf("Fehler beim Kodieren der initialen RGA: %v\n", err)
		return
	}

	// Send the size of the data first
	size := int64(bin_buf.Len())
	err = binary.Write(conn, binary.BigEndian, size)
	if err != nil {
		return
	}

	// Then send the actual data
	_, err = conn.Write(bin_buf.Bytes())
	if err != nil {
		fmt.Printf("Fehler beim Senden der initialen RGA: %v\n", err)
	}
}

func getBroadcastAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("Fehler beim Abrufen der Netzwerkinterfaces: %v", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue // Skip down and loopback interfaces
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}

			ip := ipnet.IP.To4()
			mask := ipnet.Mask

			broadcast := net.IP(make([]byte, 4))
			for i := range ip {
				broadcast[i] = ip[i] | ^mask[i]
			}

			return broadcast.String(), nil
		}
	}

	return "", fmt.Errorf("Keine geeignete IPv4-Adresse oder Broadcast-Adresse gefunden. Verfügbare Interfaces: %v", interfaces)
}

func isLocalAddress(ipToCheck string) (bool, error) {
	ip := net.ParseIP(ipToCheck)
	if ip == nil {
		return false, fmt.Errorf("ungültige IP-Adresse: %s", ipToCheck)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return false, fmt.Errorf("Fehler beim Abrufen der Netzwerkschnittstellen: %v", err)
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ipNet *net.IPNet
			var ok bool
			if ipNet, ok = addr.(*net.IPNet); !ok {
				continue
			}

			if ipNet.IP.Equal(ip) {
				return true, nil
			}
		}
	}

	return false, nil
}

func generateNetworkID() string {
	return fmt.Sprintf("network-%d", time.Now().UnixNano())
}
