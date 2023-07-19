# ngrokVNC
## Usage
- Run `ngrokVNC` when ngrok tunnel not exist it created - запусти `ngrokVNC` когда ngrok туннеля нет он создатся:
- This VNC server mode - экран VNC ожидает подключения VNC наблюдателя
- - To view over ngrok on the other side, run - для просмотра через туннель на другой стороне запусти<br>`ngrokVNC :`
- - To view over the LAN on the other side, run - для просмотра через LAN на другой стороне запусти<br>`ngrokVNC host`

- Run `ngrokVNC 0` this create ngrok tunnel - запусти `ngrokVNC 0` это создаст туннель
- This VNC viewer listen mode over ngrok - наблюдатель VNC ожидает подключения VNC экрана через тоннель
- - To view over ngrok on the other side, run - для просмотра через туннель на другой стороне запусти<br>`ngrokVNC`
- Run - запусти<br>`ngrokVNC -0`
- This VNC viewer listen mode over LAN - наблюдатель VNC ожидает подключения VNC экрана через LAN
- - To view over LAN on the other side, run - для просмотра через LAN на другой стороне запусти<br>`ngrokVNC -host`

- Run - запусти<br>`ngrokVNC -`
- This VNC server connect mode over ngrok - экран VNC ожидает туннеля VNC наблюдателя чтоб к нему подключится
- - To view over ngrok on the other side, run - для просмотра через ngrok на другой стороне запусти<br>`ngrokVNC 0`
