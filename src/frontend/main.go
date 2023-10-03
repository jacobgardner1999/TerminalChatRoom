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

    username, roomName := getUserDetails()
    setUsername(username)
    joinChatRoom(roomName)

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

	chatTitle := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Chat Room: " + roomName + "\n").
		SetDynamicColors(true).
        SetChangedFunc(func() {
			app.Draw()
		})

	chatTextView := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
        SetChangedFunc(func() {
			app.Draw()
		})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
        AddItem(chatTitle, 0, 1, false). 
		AddItem(chatTextView, 0, 3, false).
		AddItem(inputField, 3, 2, true)

    go readMessages(app, chatTextView, chatTitle)

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
}

func joinChatRoom(roomName string) { 
    sendMessage(fmt.Sprintf("/join %s", roomName))
}

func readMessages(app *tview.Application, chatTextView *tview.TextView, chatTitle *tview.TextView) {
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message: ", err)
            break
        }
        if strings.Contains(string(message), "/userRoom") {
			handleRoomUpdate(string(message), app, chatTitle)
            writeToChat("inside loop", app, chatTextView)
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
        writeToChat(m, app, chatTextView)
    }
}

func writeToChat(message string, app *tview.Application, chatTextView *tview.TextView) {
    app.QueueUpdateDraw(func() {
        chatTextView.SetText(chatTextView.GetText(true) + message)
        chatTextView.ScrollToEnd()
    })
}

func handleRoomUpdate(message string, app *tview.Application, chatTitle *tview.TextView) {
    parts := strings.Fields(message)
    if len(parts) == 2 && parts[0] == "/userRoom" {
        app.QueueUpdateDraw(func() {
            chatTitle.SetText("Chat Room: " + parts[1] + "\n")
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
