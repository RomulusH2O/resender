package main

import (
	"bufio"
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"

	"google.golang.org/grpc"

	pb "github.com/RomulusH2O/forbidden/resender/protob"
)

func main() {

	remote := "localhost"

	func() {

		file, err := os.Open("remote.txt")
		if err != nil {
			fmt.Println("Cannot open remote.txt")
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			remote = scanner.Text()
			break
		}
	}()

	fmt.Println("REMOTE: " + fmt.Sprintf("%s:9119", remote))

	uiface := &ui{}
	ctrl := &genCtrl{}

	conn, err := grpc.Dial(fmt.Sprintf("%s:9119", remote), grpc.WithInsecure())
	if err != nil {
		fmt.Println("OK")
	}

	ctrl.connection = conn
	defer ctrl.connection.Close()

	ctrl.client = pb.NewChatFileServiceClient(ctrl.connection)

	desktopApp := app.New()
	uiface.window = desktopApp.NewWindow("Resender")

	rootContent := uiface.makeUI(ctrl)

	uiface.window.SetContent(rootContent)

	ctrl.initClientId()
	ctrl.initWatchOverConnection(uiface)
	ctrl.initStayUpdatedWithServices(uiface)
	ctrl.initTrackAndShareUserTyping(uiface)

	uiface.window.ShowAndRun()
}
