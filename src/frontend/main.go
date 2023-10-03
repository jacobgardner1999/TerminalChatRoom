package main

import (
    "bufio"
	"fmt"
    "log"
	"os"
    "strings"
	"time"

	"github.com/gorilla/websocket"
)

const serverAddress = "ws://localhost:8080/ws"

var conn *websocket.Conn

type User struct {
    Username string
    Room string
} 

type Message struct {
    Sender string
    Content string
    Timestamp time.Time
}

func main() {
    fmt.Println("Welcome to the Chat CLI")

    err := initWebSocket()
    if err != nil {
        log.Fatal("Error connecting to WebSocket: ", err)
    }

    user := getUserInfo()
    joinChatRoom(user)
    go readMessages()
    handleUserInput()
}

func initWebSocket() error {
    var err error
    conn, _, err = websocket.DefaultDialer.Dial(serverAddress, nil)
    return err
}

func getUserInfo() User {
    username := getInput("Enter your username: ")

    chatRoom := getInput("Enter your chat room: ")

    return User{Username: username, Room: chatRoom}
}

func joinChatRoom(user User) { 
    sendMessage(fmt.Sprintf("/join %s %s", user.Username, user.Room))
}

func handleUserInput() {
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("Enter message or command: ")
        scanner.Scan()
        input := scanner.Text()

        if strings.HasPrefix(input, ":") {
            handleCommand(input[1:])
        } else {
            sendMessage(input)
        }
    }
}

func handleCommand(command string) {
    switch command {
    case "leave": 
        sendMessage("/leave")
    case "changename":
        newName := getInput("Enter your new name: ")
        sendMessage(fmt.Sprintf("/changename %s", newName))
    default: 
        fmt.Println("Unknown command: ", command)
    }
}

func readMessages() {
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message: ", err)
            break
        }

        parsedMessage, err := parseMessage(message)
        if err != nil {
            log.Println("Error parsing message: ", err)
            continue
        }

        fmt.Printf("[%s] %s: %s\n", parsedMessage.Timestamp.Format("HH:MM"), parsedMessage.Sender, parsedMessage.Content)
    }
}

func sendMessage(message string) {
    err := conn.WriteMessage(websocket.TextMessage, []byte(message))
    if err != nil {
        log.Println("Error sending message: ", err)
    }
}

func parseMessage(rawMessage []byte) (Message, error) {
    parts := strings.Split(string(rawMessage), "|")
    if len(parts) != 3 {
        return Message{}, fmt.Errorf("invalid message format")
    }
    sender := parts[0]
    content := parts[1]
    timestamp, err := time.Parse(time.RFC3339, parts[2])
    if err != nil {
        return Message{}, fmt.Errorf("invalid timestamp format")
    }
    return Message{Sender: sender, Content: content, Timestamp: timestamp}, nil
}

func getInput(prompt string) string {
    fmt.Print(prompt)
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    return scanner.Text()
}
