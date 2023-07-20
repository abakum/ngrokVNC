# ngrokVNC

## Credits - благодарности:

- GlavSoft - for [TightVNC](https://www.tightvnc.com)
- ngrok - for [ngrok](https://github.com/ngrok/ngrok-go)

## Usage - использование:

- git clone https://github.com/abakum/ngrokVNC
- place NGROK_AUTHTOKEN.txt and NGROK_API_KEY.txt to ngrokVNC before build or set env during run ngrokVNC
- размести NGROK_AUTHTOKEN.txt и NGROK_API_KEY.txt в ngrokVNC перед build или установи переменные окружения во время запуска ngrokVNC

- Run - запусти<br>`ngrokVNC`
- When there is no ngrok tunnel it will be created  - когда ngrok туннеля нет он создатся
- The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя
- - To view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти<br>`ngrokVNC :`
- - To view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти<br>`ngrokVNC host`

- Run - запусти<br>`ngrokVNC 0`
- This will create a ngrok tunnel - это создаст туннель
- The VNC viewer is waiting for the VNC server to connect via ngrok tunnel - наблюдатель VNC ожидает подключения VNC экрана через тоннель
- - To view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти<br>`ngrokVNC`
    
- Run - запусти<br>`ngrokVNC -0`
- The VNC viewer is waiting for the VNC server to be connected via LAN - наблюдатель VNC ожидает подключения VNC экрана через LAN
- - To view via LAN on the other side, run - для просмотра через LAN на другой стороне запусти<br>`ngrokVNC -host`

- Run - запусти<br>`ngrokVNC -`
- the VNC server is waiting for ngrok tunnel of the VNC viewer to connect to it - экран VNC ожидает туннеля VNC наблюдателя чтоб к нему подключится
- - To view over ngrok on the other side, run - для просмотра через ngrok на другой стороне запусти<br>`ngrokVNC 0`
- `ngrokVNC -` unlike - в отличии от `ngrokVNC` does not stop working when the connection is broken - не прекращает работу при разрыве соедеинения


