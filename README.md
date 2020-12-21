# Connect four multiplayer server
Go course at FMI
This is an implementation of the "Connect four" game, made for the "Introduction to Go" course at FMI. There is a server, a client with both GUI and terminal input, and you can play against a friend or AI.  
The server can handle many concurrent games as each game is in its own goroutine. It is implemented with TCP Sockets.  
There are two modes of online play:  
* Play with a friend. This works by creating a room, for which a unique token(5 character code) is generated. Other player can connect to the room with this token.  
* Quick play: You wait for a random player to connect and play with you.  
You can also play against an AI which is implemented by the alpha-beta game tree search algorithm.  
To play the game
```
cd client
./gameTerminal.sh 
or 
./gameGUI.sh
```

