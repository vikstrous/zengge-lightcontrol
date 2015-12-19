# Zengge LightControl

Zengge have a line of cheap WiFi enabled lightbulbs made in China. They don't have consistent branding, but some of the names I've seen them use in their products and apps are "magichue", "magic home", "LED magic color" and "ledmagical". They can be bought from Alibaba in bulk for anywhere between $1 and $18 depending on quantity. There are also resellers in other places under other brand names like "Flux" and "Leegoal".

Here's where they come from: [Alibaba](http://zengge.en.alibaba.com/product/1959075082-213407227/)

The bulbs use a simple binary protocol. The tools in this repository will help you use it to control the bulb over WiFi. There are also tools for controlling it over the Internet.

This command line tool and library uses what they call the "2014 protocol". The lightbulb module ID I tested with is HF-LPB100-ZJ200.

## Example usage

```
zengge-lightcontrol local --host 1.2.3.4:5577 set-power off
```

## Hardware

The main chip used is HF-LPB100. You can find out everything about it [here](http://www.hi-flying.com/products_detail/&productId=68f4fd7b-39f8-4995-93ab-53193ac5cf22.html). It's a simple WiFi access point and/or client. It runs FreeRTOS. The manufacturer provides a development board and SDK for writing applications for it.

## Software

The bulb is built using the chip manufacturer's SDK from what I could tell. It has a web server on it, but the management web UI is removed. All that's left is a server serving a basic auth prompt and then a 404 error.

Zengge has an Android and iOS app called "magic home". It has the ability to interact with both the router and the lightbulb ports (more on that later). It also has the ability to register a lightbulb with their cloud control server and control it even when not on the same LAN.

## Protocols

The bulbs have 3 ports open: TCP 80, TCP 5577 and UDP 48899. They also have the ability to make outbound connections and receive commands from a server on the Internet. Commands sent that way can control only the bulb, not the router.

The TCP port is used to control the lightbulb and the UDP port is used to control the WiFi router on it. I will be referring to the former as the bulb port and the latter as the router port. This is a very cheap embedded device, so it doesn't do any fancy cryptography or authentication. It accepts simple commands on those ports.

### HTTP server

First of all, if you want to authenticate with the HTTP server, the username/password are admin/nimda. All I see is 404 errors when I try to open it in Firefox, so not very interesting. Chrome doesn't even think it's HTTP and curl hangs, so it's not the nicest web server.

### Router Port

The router port is documented in the manual pdf [version 1.9](http://www.hi-flying.com/downloadsfront.do?method=picker&flag=all&id=dc406e44-84ec-4be1-ab11-a4ce403f6d3f&fileId=0f147d14-d0aa-4fc8-b01f-36e43418d19d), is also available.

It's running firmware version 1.0.06.

The UDP port is 48899. There is a bit of a trick to making it work. There is a "wifi configuration password" set. You have to send "HF-A11ASSISTHREAD" followed by "+ok" before you begin to interact with it / send other commands. It's a simple ASCII over UDP protocol. I think every message has to fit in one UDP frame, but I haven't confirmed that.

Many of the commands in the manual don't show up when you run AT+H, so the manual is essential. For example, `AT+SMEM` is not documented. For me it returns `12488[max_blk_size], 14692[total_size]`.

There are `AT+` commands listed in `AT+H` that differ compared to the documentation or are undocumented:

    AT+NDBGL:set/get debug level - not documented; two values - x,y; y is 0 or 1; x is int, >= 0
    AT+SLPEN: Put on/off the GPIO7. - ??? this is the Sleep_RQ pin
    AT+UPAUTO: Start the remote upgrade by config file. - invalid command error
    AT+UPCFG: Start the remote upgrade default setting. - invalid command error
    AT+UPWEB: Start the remote upgrade webpages. - invalid command error
    AT+WEBSWITCH: Set/Get the parameters of WEB page. - invalid command error
    AT+WEBVER: Get WEB version. - returns None

It seems like the device is left in its default "Transparent Transmission Mode", which means that it can receive management commands over the network instead of only over a serial port.

The module ID returned by the router is `HF-LPB100-ZJ200`
The detailed version info is `10 Cyrus.xu_Sam (2015/04/10)`
The wifi driver is 141440 bytes.

Remote control works over "HTTP protocol transfer"
The `AT+HTTPURL` command shows that the bulb connects to `http://wifi.magichue.net:8805`

Its user agent is `lwip1.3.2`

There is a strange set of commands that allow the bulb to send HTTP requests on your behalf, acting like a primitive proxy.

I found out that the router can be used in station and access point mode at the same time. This means that it's ideal for re-use as a low power simple man-in-the-middle proxy.

### Bulb Port

There are 3 modes set with `AT+TMODE`:

* Throughput mode - issue commands to lightbulb
* Command mode - ???
* PWM mode - manually adjust levels of GPIO pins

#### GPIO commands (PWM mode)

These are documented in the router documentation linked to above. I have not tried actually using them.

The router has to be set to `pwm` mode using `AT+TMODE` for these to work.

Some of the commands are:

```
GPIO <channel> OUT <value>
GPIO <channel> GET
PWM <channel> <frequency> <duty>
PWM <channel> GET
PWM <channel> SET
```

* channel can be 11,12,15,18,20,23
* value can be 1 or 0 where 1 is high voltage and 0 is low voltage
* frequency can be 500 to 60000
* duty can be 0 to 100

There are also many hex commands:

```
0a - read all GPIO channels
  -> 8a<value>
03<channel> - toggle channel value
  -> 83<channel><value>
30 - read all PWM channel frequencies
  -> b0<four two-byte values for channels 0-3>
32<channel><two byte value> - write PWM channel frequency
  -> b2<channel><two byte value>
20 - read all pwm channel duty
  -> a0 <four bytes for channels 0-3>
24... - write all PWM channel duty
  -> a4...
22... - write PWM channel duty
  -> a2...
71 - save present GPIO,PWM settings
  -> fa
04 - assert all GPIO channels low
  -> 8400
05 - assert all GPIO channels high
  -> 85 01
7e - read resources of module
  -> fe<output pin><input pin><pwm pin>
```

#### Bulb commands (throughput mode)

This protocol is less well documented. It's a binary protocol. All examples I'll be showing are in hex. I think my code is better documentation than this README can provide.

All commnads are of fixed length. The last byte is always a checksum.  The checksum is just the sum of all the previous bytes in the current command.

The following examples can give you an idea of how the protocol works. The rest can be seen re-implemented in protocol.go.

This example command **changes the color**:

```
31RRGGBBWWXXCC
```

where:

* 31 is the ID of the command
* RR is red
* GG is green
* BB is blue
* WW is white
* XX is whether to use the white value or not
* CC is the checksum

Booleans such as XX are represented as 0xf0 for True and 0x0f for False.
There is no response.

This command queries the **state of the bulb**:

```
818a8b96
```

The result looks like this:

```
814423612101fefb8a000400f0e2
```

It includes the following information:

```go
type State struct {
	DeviceType    uint8
	IsOn          bool
	LedVersionNum uint8
	Mode          uint8
	Slowness      uint8
	Color         Color
}
```

### Cloud Control

This is called "remote" in the Android app. It allows the bulb to be controlled over the Internet.

The connection is established this way:

1. The phone finds the mac address of the bulb
1. The phone "logs in" to the cloud server using its device id and gets a cookie
1. The phone registers the bulb by telling the server its mac address. This associates the bulb with the device id / user
1. The phone uses AT+ commands to tell the bulb to connect to the cloud server
1. The bulb makes a connection and tells the server its mac address

Commands are sent to the bulb like this:

1. The phone logs into the cloud server with its device id and gets a cookie
1. The phone can now send commands to the bulb using the bulb binary protocol tunneled over HTTP to the server and over TCP from the server to the bulb

Device ids are uuid4 converted into numbers on android and in uuid format on iOS. In practice they can be any unique string.

All requests to the cloud server from the phone use "authentication" which consists of a simple AES256 encryption with a shared secret distributed with the application.

Every api call is sent to `http://wifi.magichue.net/WebMagicHome/ZenggeCloud/ZJ002.ashx` (yes, unencrypted). The body of the request is url encoded with content type `application/x-www-form-urlencoded`. The command to execute is in a header called `zg-app-cmd`.

Example login request:

```
POST /WebMagicHome/ZenggeCloud/ZJ002.ashx HTTP/1.1
Content-Type: application/x-www-form-urlencoded;charset=UTF-8
Accept-Charset: UTF-8
zg-app-cmd: Login
Host: wifi.magichue.net
Accept-Encoding: gzip
Content-Length: 439

AppKey=65ee4e302f844df87939cbe879041a7ba2d0df17&DevID=<device id redacted>&AppVer=1.0.9&CheckCode=c2+50+c9+6e+5f+76+f3+b2+1a+76+a1+c7+25+9d+d8+cc+53+b0+e6+ff+45+49+15+f6+72+30+28+c4+d9+f7+f6+29+e8+3e+04+3b+40+12+b5+12+75+eb+a7+3b+57+e6+ee+e9+af+54+79+d5+ac+af+b2+ca+7d+8d+ea+3d+1a+eb+c7+6f+31+0a+a6+db+55+e3+d8+f6+b2+27+74+c3+f3+65+f5+8d+2c+b6+d5+b3+6e+f4+ac+8f+bf+44+58+90+a8+3c+83+99+&AppSys=Android&Timestamp=Sat+Aug+08+17%3A34%3A01+PDT+2015
```
