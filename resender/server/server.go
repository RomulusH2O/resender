package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	misc "github.com/RomulusH2O/forbidden/resender/misc"
	pb "github.com/RomulusH2O/forbidden/resender/protob"
)

const (
	DEFAULT_SERVER_IP   = "localhost"
	DEFAULT_SERVER_PORT = "9119"
	DEFAULT_SERVER_TYPE = "tcp"
)

var (
	ctrl *genCtrl
)

type genCtrl struct {
	chatCtrl
	filesCtrl
	dbCtrl
}

type chatCtrl struct {
	chatUpdateInstantTicker       *time.Ticker
	userTypingUpdateInstantTicker *time.Ticker
	typingUsers                   map[string]bool
	typingUsersLock               sync.RWMutex
	// chatMessages                  []string
}

type filesCtrl struct {
	knownFilesUpdateInstantTicker *time.Ticker
	knownFiles                    []string
}

type dbCtrl struct {
	database *gorm.DB
}

type Server struct {
	pb.ChatFileServiceServer
}

func main() {

	fmt.Printf("Setting up database...\n")

	db, err := gorm.Open(sqlite.Open("./sqlitedb/chat.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database")
	}
	fmt.Printf("The database is OK\n")

	ctrl = &genCtrl{}

	ctrl.database = db

	ctrl.database.AutoMigrate(&ChatMessageEntity{})

	ctrl.initStayPreparedForService()

	serviceIP := DEFAULT_SERVER_IP
	servicePort := DEFAULT_SERVER_PORT

	flag.StringVar(&serviceIP, "service-ip", DEFAULT_SERVER_IP, "service_ip")
	flag.StringVar(&servicePort, "service-port", DEFAULT_SERVER_PORT, "service_port")

	flag.Parse()

	if isValidIP := misc.IsValidIP4(serviceIP); !isValidIP {
		fmt.Printf("Did not provide a valid IP address (will use: %s)\n", DEFAULT_SERVER_IP)
	} else {
		fmt.Printf("Service listen IP address was provided (will use: %s)\n", serviceIP)
	}

	fmt.Printf("Service listen IP set to: %s\n", serviceIP)
	fmt.Printf("Service listen port set to: %s\n", servicePort)

	fmt.Printf("Starting server...\n")

	serviceIPPort := fmt.Sprint(serviceIP, ":", servicePort)
	lis, err := net.Listen(DEFAULT_SERVER_TYPE, serviceIPPort)

	if err != nil {
		log.Fatalf("Failed to listen on %v\n", err)
	}
	log.Printf("Listening on %s\n", serviceIPPort)

	server := grpc.NewServer()
	pb.RegisterChatFileServiceServer(server, &Server{})

	if err = server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v\n", err)
	}
}

func (c *genCtrl) initStayPreparedForService() {

	c.knownFilesUpdateInstantTicker = time.NewTicker(time.Second * 3)
	c.knownFiles = []string{}
	go c.stayPreparedForKnownFiles()

	c.chatUpdateInstantTicker = time.NewTicker(time.Second * 3)
	go c.stayPreparedForChat()

	c.typingUsers = make(map[string]bool, 256)
	c.userTypingUpdateInstantTicker = time.NewTicker(time.Second * 6)
	go c.stayPreparedForUserTyping()
}

func (c *genCtrl) stayPreparedForKnownFiles() {

	for {
		<-c.knownFilesUpdateInstantTicker.C

		entries, errReadDir := os.ReadDir("./uploaded")

		if errReadDir != nil {
			fmt.Println(errReadDir.Error())
			return
		}

		knownFilesTmp := []string{}

		for _, entry := range entries {

			if !entry.IsDir() {

				knownFilesTmp = append(knownFilesTmp, entry.Name())
			}
		}
		c.knownFiles = knownFilesTmp
	}
}

func (c *genCtrl) stayPreparedForChat() {

	for {
		<-c.chatUpdateInstantTicker.C

	}
}

func (c *genCtrl) stayPreparedForUserTyping() {

	for {
		<-c.userTypingUpdateInstantTicker.C

		/*for k := range TypingUsers {

			TypingUsers[k] = false
		}*/
		c.typingUsersLock.RLock()

		maps.Clear(c.typingUsers)

		c.typingUsersLock.RUnlock()
	}
}
