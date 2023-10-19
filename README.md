# ngrokVNC
Helps VNC viewer communicate with a VNC server over firewalls via ngrok.com<BR>
Помогает VNC наблюдателю просматривать VNC экран через фаерволы используя ngrok.com


## Credits - благодарности:

- GlavSoft - for [TightVNC](https://www.tightvnc.com)
- Rudi De Vos, Sam Liarfo, Ludovic Bocquet -for [UltraVNC](https://uvnc.com/)
- ngrok - for [ngrok](https://github.com/ngrok/ngrok-go)

## Usage - использование:

- `git clone https://github.com/abakum/ngrokVNC`
- place - размести `NGROK_AUTHTOKEN.txt` and - и `NGROK_API_KEY.txt` to - в `ngrokVNC` before - перед `go build .` or set env during run - или установи переменные окружения во время запуска `ngrokVNC`
- Run - запусти<br>`ngrokVNC [::port]`
- When there is no ngrok tunnel it will be created  - когда ngrok туннеля нет он создатся
- The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя
- - To view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти<br>`ngrokVNC :`
- - To view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти<br>`ngrokVNC host[::port]`

- Run - запусти<br>`ngrokVNC 0`
- This will create a ngrok tunnel - это создаст туннель
- The VNC viewer is waiting for the VNC server to connect via ngrok tunnel - наблюдатель VNC ожидает подключения VNC экрана через тоннель
- - To view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти<br>`ngrokVNC [::port]`
    
- Run - запусти<br>`ngrokVNC -0`
- The VNC viewer is waiting for the VNC server to be connected via LAN - наблюдатель VNC ожидает подключения VNC экрана через LAN
- - To view via LAN on the other side, run - для просмотра через LAN на другой стороне запусти<br>`ngrokVNC -host`

- Run - запусти<br>`ngrokVNC -`
- the VNC server is waiting for ngrok tunnel of the VNC viewer to connect to it - экран VNC ожидает туннеля VNC наблюдателя чтоб к нему подключится
- - To view over ngrok on the other side, run - для просмотра через ngrok на другой стороне запусти<br>`ngrokVNC 0`

[Распутать запутанные параметры](args.txt)


