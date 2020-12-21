# Connect four multiplayer server
Go course at FMI
This is an implementation of the "Connect four" game, made for the "Introduxction to Go" course at FMI.
There is a server, a client with both GUI and terminal input, and you can play against a friend or AI.  
The server can hadle many concurrent games as each game is in its own gorutine.  
There are two modes of online play:  
-Play with with a friend. This works by creating a room, for which a unique token is generated. Other player can connect ot the room with this token.  
-Quick play: You wait for a random palyer to connect and play with you.  
You can also play against an AI which is implemented by the alpha-beta game tree search algorithm.  
To play the game
```
cd client
./gameTerminal.sh 
or 
./gameGUI.sh
```

