package network

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

const (
    broadcastAddr = "255.255.255.255:12345"
    tcpBasePort   = 12346
)

type Session struct {
    Name string
    IP   string
    Port int
}

var (
    Sessions     = make(map[string]Session)
    sessionMutex sync.Mutex
)

func Test() {
    go ListenForBroadcasts()
    go BroadcastSession()

    time.Sleep(5 * time.Second) // Warte kurz, um einige Sitzungen zu entdecken

    for {
        sessionMutex.Unlock()

        fmt.Println("\nOptionen:")
        fmt.Println("1. Sitzung beitreten")
        fmt.Println("2. Aktualisieren")
        fmt.Println("3. Beenden")

        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Wählen Sie eine Option: ")
        choice, _ := reader.ReadString('\n')
        choice = strings.TrimSpace(choice)

        switch choice {
        case "1":
            fmt.Print("Geben Sie den Namen der Sitzung ein, der Sie beitreten möchten: ")
            sessionName, _ := reader.ReadString('\n')
            sessionName = strings.TrimSpace(sessionName)
            JoinSession(sessionName)
        case "2":
            continue
        case "3":
            return
        default:
            fmt.Println("Ungültige Option, bitte versuchen Sie es erneut.")
        }
    }
}

func ListenForBroadcasts() (map[string]Session) {
    addr, err := net.ResolveUDPAddr("udp", broadcastAddr)
    if err != nil {
        fmt.Println("Fehler beim Auflösen der Broadcast-Adresse:", err)
        return nil
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Println("Fehler beim Öffnen der UDP-Verbindung:", err)
        return nil
    }
    defer conn.Close()

    buffer := make([]byte, 1024)
    for {
        n, remoteAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Fehler beim Lesen der UDP-Nachricht:", err)
            continue
        }

        message := string(buffer[:n])
        parts := strings.Split(message, "|")
        if len(parts) == 3 && parts[0] == "SESSION" {
            sessionName := parts[1]
            port, _ := strconv.Atoi(parts[2])
            sessionMutex.Lock()
            Sessions[sessionName] = Session{Name: sessionName, IP: remoteAddr.IP.String(), Port: port}
            sessionMutex.Unlock()
        }
    }
}

func BroadcastSession() {
    addr, err := net.ResolveUDPAddr("udp", broadcastAddr)
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

    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        fmt.Println("Fehler beim Starten des TCP-Listeners:", err)
        return
    }
    defer listener.Close()

    port := listener.Addr().(*net.TCPAddr).Port
    sessionName := fmt.Sprintf("Session-%d", port)
    message := []byte(fmt.Sprintf("SESSION|%s|%d", sessionName, port))

    go func() {
        for {
            _, err := conn.Write(message)
            if err != nil {
                fmt.Println("Fehler beim Senden der Broadcast-Nachricht:", err)
            }
            time.Sleep(5 * time.Second)
        }
    }()

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Fehler beim Akzeptieren der Verbindung:", err)
            continue
        }
        go HandlePeer(conn)
    }
}

func JoinSession(sessionName string) {
    sessionMutex.Lock()
    session, exists := Sessions[sessionName]
    sessionMutex.Unlock()

    if !exists {
        fmt.Println("Sitzung nicht gefunden.")
        return
    }

    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", session.IP, session.Port))
    if err != nil {
        fmt.Println("Fehler beim Verbinden mit der Sitzung:", err)
        return
    }
    defer conn.Close()

    fmt.Printf("Verbunden mit Sitzung %s\n", sessionName)
    HandlePeer(conn)
}

func HandlePeer(conn net.Conn) {
    defer conn.Close()
    fmt.Printf("Verbunden mit Peer: %s\n", conn.RemoteAddr())

    reader := bufio.NewReader(os.Stdin)
    netReader := bufio.NewReader(conn)

    for {
        fmt.Print("Nachricht eingeben (oder 'exit' zum Beenden): ")
        message, _ := reader.ReadString('\n')
        message = strings.TrimSpace(message)

        if message == "exit" {
            return
        }

        _, err := conn.Write([]byte(message + "\n"))
        if err != nil {
            fmt.Println("Fehler beim Senden der Nachricht:", err)
            return
        }

        response, err := netReader.ReadString('\n')
        if err != nil {
            fmt.Println("Fehler beim Lesen der Antwort:", err)
            return
        }

        fmt.Printf("Antwort erhalten: %s", response)
    }
}
