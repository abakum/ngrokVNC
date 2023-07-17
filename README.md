# ngrokVNC
## Usage
- Run `ngrokVNC` when ngrok not exist - запусти `ngrokVNC` когда ngrok туннель отсутствует:
- - VNC server mode - экран VNC ожидает подключения VNC наблюдателя
- - - To view over ngrok on the other side, run - для просмотра через туннель на другой стороне запусти<br>`ngrokVNC :`
- - - To view over the LAN on the other side, run - для просмотра через LAN на другой стороне запусти<br>`ngrokVNC host`
- Run `ngrokVNC` when ngrok exist - запусти `ngrokVNC` когда ngrok туннель создан:
- - VNC server connect to viewer mode - экран VNC подключается к ожидающему наблюдателю
- - On the other side was launched - на другой стороне был запущен<br>`ngrokVNC 0`
- To connect the VNC server to the viewer over the ngrok, run - для подключения экрана VNC к наблюдателю через туннель запусти<br>`ngrokVNC ::`
- - On the other side run - на другой стороне запусти<br>`ngrokVNC 0`
