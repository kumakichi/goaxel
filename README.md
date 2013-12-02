goaxel
======
Goroutines Download Accelerator

install 
=======
```
go get github.com/xiangzhai/goaxel 
go get github.com/cheggaaa/pb
```

usage 
=====
* go run goaxel.go -d http://www.linuxdeepin.com/index.en.html 
* go run goaxel.go -d -n 6 http://www.linuxdeepin.com/index.en.html

screenshot 
==========
![ScreenShot](https://raw.github.com/xiangzhai/goaxel/master/doc/console.png)

TODO 
====
- [x] command parse args
- [x] HTTP protocol
- [x] FTP protocol
- [x] continue downloading from the point of interruption
- [x] command progress bar
- [x] mirror support
- [x] HTTPS protocol
- [ ] BitTorrent protocol
- [ ] eMule protocol
- [ ] MagNet protocol
- [ ] QML UI
