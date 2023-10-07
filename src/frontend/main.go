package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
)

const serverAddress = "ws://localhost:8080/ws"

var conn *websocket.Conn

type Message struct {
    Sender string
    Content string
    Timestamp string
}
    
var app = tview.NewApplication()
var chatTitle = tview.NewTextView()
var chatTextView = tview.NewTextView()

func main() {
    fmt.Println("Welcome to the Chat CLI")

    err := initWebSocket()
    if err != nil {
        log.Fatal("Error connecting to WebSocket: ", err)
    }

    username, roomName := getUserDetails()
    setUsername(username)
    joinChatRoom(roomName)

	inputField := tview.NewInputField()
    inputField.SetLabel("Type your message: ").
        SetFieldWidth(50).
        SetFieldBackgroundColor(tcell.ColorBlack).
        SetDoneFunc(func(key tcell.Key) {
            if key == tcell.KeyEnter {
                sendMessage(inputField.GetText())
                inputField.SetText("")
            }
        })

	chatTitle.SetTextAlign(tview.AlignCenter).
		SetText("Chat Room: " + roomName + "\n").
		SetDynamicColors(true).
        SetChangedFunc(func() {
			app.Draw()
		})

	chatTextView.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
        SetChangedFunc(func() {
			app.Draw()
		})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
        AddItem(chatTitle, 0, 1, false). 
		AddItem(chatTextView, 0, 9, false).
		AddItem(inputField, 3, 2, true)

    go readMessages()

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatal("Error running application: ", err)
	}
}

func initWebSocket() error {
    var err error
    conn, _, err = websocket.DefaultDialer.Dial(serverAddress, nil)
    return err
}

func getUserDetails() (username string, roomName string) {
    name := getInput("Enter your username: ")

    chatRoom := getInput("Enter your chat room: ")

    return name, chatRoom
}

func setUsername(username string) {
    sendMessage(fmt.Sprintf("/name %s", username))
    time.Sleep(5 * time.Millisecond)
}

func joinChatRoom(roomName string) { 
    sendMessage(fmt.Sprintf("/join %s", roomName))
    time.Sleep(5 * time.Millisecond)
}

func readMessages() {
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message: ", err)
            break
        }
        if strings.Contains(string(message), "/userRoom") {
			handleRoomUpdate(string(message))
			continue
		}
        parsedMessage, err := parseMessage(message)
        if err != nil {
            log.Println("Error parsing message: ", err)
            continue
        }
        m := fmt.Sprintf("[%s] %s: %s\n",
        parsedMessage.Timestamp,
        parsedMessage.Sender,
        parsedMessage.Content)
        writeToChat(m)
    }
}

func writeToChat(message string) {
    app.QueueUpdateDraw(func() {
        chatTextView.SetText(chatTextView.GetText(true) + message)
        chatTextView.ScrollToEnd()
    })
}

func handleRoomUpdate(message string) {
    parts := strings.Fields(message)
    if parts[len(parts)-2] == "/userRoom" {
        app.QueueUpdateDraw(func() {
            chatTitle.SetText("Chat Room: " + parts[len(parts)-1] + "\n")
            chatTextView.SetText("")
        })
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
        return Message{}, fmt.Errorf("invalid message format. Expected length 3, got length" + fmt.Sprint(len(parts)) + "\n" + string(rawMessage))
    }
    sender := parts[0]
    timestamp := parts[1]
    content := parts[2]
    return Message{Sender: sender, Content: content, Timestamp: timestamp}, nil
}

func getInput(prompt string) string {
    fmt.Print(prompt)
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    return scanner.Text()
}
