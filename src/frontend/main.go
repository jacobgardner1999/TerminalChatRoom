package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
    "github.com/rivo/tview"
    "github.com/gdamore/tcell/v2"
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
    Timestamp string
}

func main() {
    fmt.Println("Welcome to the Chat CLI")

    err := initWebSocket()
    if err != nil {
        log.Fatal("Error connecting to WebSocket: ", err)
    }

    user := getUserInfo()
    setUsername(user.Username)
    joinChatRoom(user)

    app := tview.NewApplication()

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

	chatTextView := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText("Chat Room: " + user.Room + "\n").
		SetDynamicColors(true).
        SetChangedFunc(func() {
			app.Draw()
		})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chatTextView, 0, 1, false).
		AddItem(inputField, 3, 1, true)

    go readMessages(app, chatTextView)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatal("Error running application: ", err)
	}
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

func setUsername(username string) {
    sendMessage(fmt.Sprintf("/name %s", username))
}

func joinChatRoom(user User) { 
    sendMessage(fmt.Sprintf("/join %s", user.Room))
}

func readMessages(app *tview.Application, chatTextView *tview.TextView) {
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

        app.QueueUpdateDraw(func() {
            chatTextView.SetText(chatTextView.GetText(true) + fmt.Sprintf("[%s] %s: %s\n",
            parsedMessage.Timestamp,
            parsedMessage.Sender,
            parsedMessage.Content))
            chatTextView.ScrollToEnd()
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
        return Message{}, fmt.Errorf("invalid message format")
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
