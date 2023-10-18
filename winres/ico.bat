chcp 65001
set c=blue
set t=VnC
set x=(w-text_w)/2
set o=icon32
set s=32
set f=18
call :b

set t=ngrokVNC
set o=icon
set s=256
set f=60
call :b
goto :EOF


:b
ffmpeg -f lavfi -i color=%c%:size=%s%x%s% -frames:v 1 -filter_complex drawtext=text=%t%:font=impact:fontcolor=white:fontsize=%f%:x=%x%:y=(h-text_h)/2:shadowx=3:shadowy=2 -y %o%.png
goto :EOF

