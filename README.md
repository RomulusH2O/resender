# Resender
Simple chat desktop app with client and server.
Communicate with people over the network.

Features:
- Send and view text messages
- Load earlier text messages/attachments
- Upload/download/list attachments
- Set nickname (users also get unique IDs)

Additional features:
- Once connections is lost, Client will always attempt to reconnect (and notify the user)
- Show attachment upload progress (only the uploader)
- Tell the user other users are now typing


![Zrzut ekranu 2024-12-20 163118](https://github.com/user-attachments/assets/c67e1189-da20-48fd-b07c-e1fc7e18e834)


Yet to be implemented:
- Pop-up notifications on new messages
- Search attachments (button is inactive)

Not planned:
- Login/register system for users
- Multiple chats (not just one global chat)

To be fixed:
- Keep original names of the sent attachments
- The "Other users are typing..." feature should be more responsive

Tested on platforms:
- Linux (Ubuntu)
- Windows (11)


How to build

1. Clone the repository.
   ```sh
   git clone https://github.com/github_username/repo_name.git](https://github.com/RomulusH2O/resender
   ```
2. Compile the Client.
- Linux:
  ```sh
  go build
   ```
- Windows:
   ```sh
   GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -ldflags -H=windowsgui
   ```
3. Compile the Server
- Linux:
  ```sh
  go build
   ```
- Windows:
   ```sh
   GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build
   ```


How to run

Run the Server
   ```sh
   ./server -service-ip YOUR.SUBNET.IPv4.ADDRESS
   ```
Run the Client(s)
- First set the Server's IP address in the file remote.txt
   ```sh
   ./resender
   ```

  
