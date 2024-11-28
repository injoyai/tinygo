### 说明

1. 使用Pico (`测试通过`)
   * 用安卓线连接Pico,Pico按`boot`键,然后连接电脑
   * 使用命令`tinygo flash -target=pico main.go`进行烧录
   * 烧录完成即直接运行该程序

2. 使用ESP8266 (`待测试`)
编译命令 `tinygo build -target=esp8266 -o main.bin  main.go`
烧录命令 `tinygo flash -target=esp8266 -port=COM2  main`

3. STM32命令 (`待测试`)
   * 下载烧录软件FlyMcu `http://www.mcuisp.com/software/FlyMcu.rar`
   * 编译命令 `tinygo build -target=stm32f4disco -o main.hex main.go`
   * 连接`USB转TTL`到STM32板子上,VCC->3.3或3V3,GND->GND(任意),RXD->A9,RXT->A10
   * 修改跳线位置,并按下复位按钮
   * `联机下载时的程序文件`中选择编译好的hex文件
   * 按下`FlyMcu`的`开始编程`按钮,等待烧录完成
   * 修改跳线位置,并按下复位键

### 常见问题

1. error: failed to flash C:\Users\USER\AppData\Local\Temp\tinygo716933404\main.hex: exec: "openocd": executable file not found in %PATH%
   解决方法: `openocd`下载地址`https://github.com/openocd-org/openocd/releases/download/latest/openocd-133dd9d66-i686-w64-mingw32.tar.gz`
