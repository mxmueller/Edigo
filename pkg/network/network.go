package network

import (
	"bytes"
	"edigo/pkg/crdt"
	"encoding/gob"
	"fmt"
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
    broadcastMin = 12340
    broadcastMax = 12344
    tcpBasePort   = 12346
)

type Network struct{ // das state managment is ehrenlos
    IsHost bool
    Host net.Conn // isHost = flase
    Clients []net.Conn // isHost = true
    Sessions     map[string]Session // found connections
    NewConnection chan net.Conn
    CurrentSession string // "" -> keine Session offen
    UdpPort int
}

type Session struct {
    Name string
    IP   string
    Port int
}

var (
    sessionMutex sync.Mutex
    Connections []net.Conn // active connections
    Quit = make(chan bool)
)

func NewNetwork() *Network{

    broadcastAddr, err := getBroadcastAddress()
	if err != nil {
		fmt.Println("Fehler:", err)
        os.Exit(1)
	}
    
    udpPort := 0

    for port := broadcastMin; port <= broadcastMax; port++ {
        addr, err := net.ResolveUDPAddr("udp", broadcastAddr + ":" + strconv.Itoa(port))
        if err != nil {
            fmt.Println("Fehler beim Auflösen der Broadcast-Adresse:", err)
            continue
        }

        conn, err := net.ListenUDP("udp", addr)
        if err != nil {
            continue
        }
        udpPort = port
        defer conn.Close()
        break
    }
    if udpPort == 0 {
        fmt.Println("Kein freier Port mehr verfügbar. Stelle eine größere Portrange ein")
        os.Exit(1)
    }

    return &Network{IsHost: false, Sessions: make(map[string]Session), UdpPort: udpPort, CurrentSession: ""}
}

func (network *Network) ListenForBroadcasts()  {
    addr, err := net.ResolveUDPAddr("udp", broadcastAddr + ":" + strconv.Itoa(network.UdpPort))
    if err != nil {
        fmt.Println("Fehler beim Auflösen der Broadcast-Adresse:", err)
        return 
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Println("Fehler beim Öffnen der UDP-Verbindung:", err)
        return 
    }

    buffer := make([]byte, 1024)
    for {
        n, remoteAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Fehler beim Lesen der UDP-Nachricht:", err)
            continue
        }
        isLocalAddr, err := isLocalAddress(remoteAddr.IP.String())

        message := string(buffer[:n])
        parts := strings.Split(message, "|")
        
        udpPort, err := strconv.Atoi(parts[3])
        
        // fmt.Println(parts[3], strconv.Itoa(network.UdpPort))
        if isLocalAddr && udpPort == network.UdpPort{
            continue
        }

        if len(parts) == 4 && parts[0] == "SESSION" {
            sessionName := parts[1]
            port, _ := strconv.Atoi(parts[2])
            sessionMutex.Lock()
            network.Sessions[sessionName] = Session{Name: sessionName, IP: remoteAddr.IP.String(), Port: port}
            sessionMutex.Unlock()
        }
    }
}

func (network *Network) BroadcastSession(rga *crdt.RGA) {

    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        fmt.Println("Fehler beim Starten des TCP-Listeners:", err)
        return
    }

    port := listener.Addr().(*net.TCPAddr).Port
    sessionName := fmt.Sprintf("Session-%d", port)

    go func() {
        for {
            select {
                case <- Quit:
                return
            default:

                for udpPort := broadcastMin; udpPort <= broadcastMax; udpPort ++{

                    message := []byte(fmt.Sprintf("SESSION|%s|%d|%d", sessionName, port, network.UdpPort))
                    addr, err := net.ResolveUDPAddr("udp", broadcastAddr + ":" + strconv.Itoa(udpPort))
                    if err != nil {
                        fmt.Println("Fehler beim Auflösen der Broadcast-Adresse:", err)
                        return
                    }

                    conn, err := net.DialUDP("udp", nil, addr)
                    if err != nil {
                        fmt.Println("Fehler beim Öffnen der UDP-Verbindung:", err)
                        return
                    }
                    defer conn.Close()
                    _, err = conn.Write(message)
                    if err != nil {
                        fmt.Println("Fehler beim Senden der Broadcast-Nachricht:", err)
                    }
                }
                    time.Sleep(3 * time.Second)
                }
            }
        }()

    for {
        select {
            case <- Quit:
            return
        default:
            conn, err := listener.Accept()
            if err != nil {
                fmt.Println("Fehler beim Akzeptieren der Verbindung:", err)
                continue
            }
                network.Clients = append(network.Clients, conn) // neuer client
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
        fmt.Println("Fehler beim Verbinden mit der Sitzung:", err)
        return crdt.RGA{}
    }

    tmp := make([]byte, 2024)
    _, err = conn.Read(tmp)
    tmpbuff := bytes.NewBuffer(tmp)

    tmpstruct := new(crdt.RGA)

    gobobj := gob.NewDecoder(tmpbuff)

    gobobj.Decode(tmpstruct)
    
    network.Host = conn
    network.NewConnection <- conn
    network.IsHost = false
    network.CurrentSession = session.Name

    return *tmpstruct
}

func (network *Network) SendOperation(op crdt.Operation, conn net.Conn){

    bin_buf := new(bytes.Buffer)
    gobobj := gob.NewEncoder(bin_buf)
    gobobj.Encode(op)
    conn.Write(bin_buf.Bytes())

}

func (network *Network) CloseAsHost(){
    
    for _, conn := range network.Clients{
        conn.Close()
    }
    network.CurrentSession = ""
    network.IsHost = false
    
}

func (network *Network) HostClosedSession(){
    network.Host = nil
    network.CurrentSession = ""
}

func (network *Network) RemoveClient(conn net.Conn){
    for i, c := range network.Clients {
        if c == conn {
            network.Clients = append((network.Clients)[:i], (network.Clients)[i+1:]...)
            return
        }
    }
}

func SendInitRGA(rga crdt.RGA, conn net.Conn){

    bin_buf := new(bytes.Buffer)
    gobobj := gob.NewEncoder(bin_buf)
    gobobj.Encode(rga)
    conn.Write(bin_buf.Bytes())
}

func getRandomPort(min, max int) int {
	return rand.Intn(max-min+1) + min
}


func getBroadcastAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("Fehler beim Abrufen der Netzwerkinterfaces: %v", err)
	}

	for _, iface := range interfaces {
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

	return "", fmt.Errorf("Keine geeignete IPv4-Adresse gefunden")
}

func isLocalAddress(ipToCheck string) (bool, error) {
	// Wandelt die Eingabe-IP in ein net.IP-Objekt um
	ip := net.ParseIP(ipToCheck)
	if ip == nil {
		return false, fmt.Errorf("ungültige IP-Adresse: %s", ipToCheck)
	}

	// Holt alle Netzwerkschnittstellen
	interfaces, err := net.Interfaces()
	if err != nil {
		return false, fmt.Errorf("Fehler beim Abrufen der Netzwerkschnittstellen: %v", err)
	}

	// Iteriere über alle Interfaces und überprüfe die IP-Adressen
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Prüfen, ob es sich um eine IP-Adresse handelt
			var ipNet *net.IPNet
			var ok bool
			if ipNet, ok = addr.(*net.IPNet); !ok {
				continue
			}

			// Überprüfen, ob die IP-Adresse zu diesem Interface gehört
			if ipNet.IP.Equal(ip) {
				return true, nil
			}
		}
	}

	// IP-Adresse wurde nicht gefunden
	return false, nil
}
