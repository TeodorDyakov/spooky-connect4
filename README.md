# Connect four multiplayer server
This is an implementation of the [Connect four](https://en.wikipedia.org/wiki/Connect_Four) game, made for the "Introduction to Go" course at FMI. There is a server, a client with a GUI and you can play against a friend or AI. For the GUI I have used the [Ebiten](https://github.com/hajimehoshi/ebiten) game engine.

![screen](client/images/screen.jpg)

There are two modes of online play:  
* Play with a friend. This works by creating a room, for which a unique token(5 character code) is generated. Other player can connect to the room with this token.  
* Quick play: You wait for a random player to connect and play with you.  

The server can handle many concurrent games as each game is in its own goroutine. It is implemented with TCP Sockets. 
You can also play against an AI which is implemented by the alpha-beta game tree search algorithm.  

To build and run locally:
```
git clone https://github.com/TeodorDyakov/spooky-connect4
cd spooky-connect4/client
go run -x main.go
```
To start a server, on given port:
```
go build server.go
./server -port=12345
```
To change the server to which the client connects ```cd client``` and edit the ```connectionConfig``` file (default is localhost:12345).
