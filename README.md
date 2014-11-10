goaxel
======
Goroutines Download Accelerator

install 
=======
```
go get github.com/kumakichi/goaxel 
go get github.com/cheggaaa/pb
go build
```

usage 
=====
* goaxel -d http://www.linuxdeepin.com/index.en.html 
* goaxel -n 6 http://www.linuxdeepin.com/index.en.html

screenshot 
==========
![ScreenShot](https://raw.github.com/kumakichi/goaxel/master/doc/console.png)

TODO 
====
- [x] command parse args
- [x] HTTP protocol
- [x] FTP protocol
- [x] continue downloading from the point of interruption
- [x] command progress bar
- [x] mirror support
- [x] HTTPS protocol
- [ ] recursive
- [ ] convert links
- [ ] page requisites
- [ ] reject list
- [ ] BitTorrent protocol
- [ ] eMule protocol
- [ ] MagNet protocol
- [ ] QML UI
