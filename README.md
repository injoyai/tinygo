### 说明

1. ESP8266命令
编译命令 `tinygo build -target=esp8266 -o main.bin  main.go`
烧录命令 `tinygo flash -target=esp8266 -port=COM2  main`

2. STM32命令
   1. 下载烧录软件FlyMcu `http://www.mcuisp.com/software/FlyMcu.rar`
   2. 编译命令 `tinygo build -target=stm32f4disco -o main.hex main.go`
   3. 连接`USB转TTL`到STM32板子上,VCC->3.3或3V3,GND->GND(任意),RXD->A9,RXT->A10
   4. 修改跳线位置,并按下复位按钮
   5. `联机下载时的程序文件`中选择编译好的hex文件
   6. 按下`FlyMcu`的`开始编程`按钮,等待烧录完成
   7. 修改跳线位置,并按下复位键

### 常见问题

1. error: failed to flash C:\Users\USER\AppData\Local\Temp\tinygo716933404\main.hex: exec: "openocd": executable file not found in %PATH%
   解决方法: `openocd`下载地址`https://github.com/openocd-org/openocd/releases/download/latest/openocd-133dd9d66-i686-w64-mingw32.tar.gz`
