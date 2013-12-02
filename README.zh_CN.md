goaxel
======
Goroutines 下载加速器

安装 
====
go get github.com/xiangzhai/goaxel 

go get github.com/cheggaaa/pb

使用 
====
* go run goaxel.go http://www.linuxdeepin.com/index.en.html 
* go run goaxel.go -d -n 6 http://www.linuxdeepin.com/index.en.html

截图 
====
![ScreenShot](https://raw.github.com/xiangzhai/goaxel/master/doc/console.png)

TODO 
====
- [x] 命令行下解析参数
- [x] HTTP协议支持
- [x] FTP协议支持
- [x] 断点续传
- [x] 命令行下进度条
- [x] 镜像支持
- [ ] HTTPS协议支持
- [ ] BitTorrent协议支持
- [ ] eMule协议支持
- [ ] MagNet协议支持
- [ ] QML用户界面

项目来由 
========
最近在学习GoLang, 想通过做一个小项目来上手, 因为在Linux Deepin我写过Axel的
Python绑定 https://github.com/linuxdeepin/deepin-axel 
也把Axel移植到WIN32上 http://www.codeproject.com/Articles/335690/MultiThread-Download-Accelerator-Console 
所以, 这次选择GoAxel项目, 完全用GoLang来替代C语言写的Axel. 

GoAxel不完全遵循Axel的设计思路, 毕竟Axel是上个世纪开发的, 但是都抽象出了conn
数据结构, conn根据HTTP/FTP协议, 动态选择使用http/ftp处理下载. GoAxel使用了传说
中比线程廉价的Goroutines, 为了实现"多协程"的断点续传, GoAxel将文件分成多个块儿, 
当断点续传的时候, 根据块儿的个数, 来决定重新发起多少个conn, 根据每个块儿的文件
大小, 来决定每个conn的偏移量.
 
GoAxel是真正的并行读socket/写块儿文件, 而Axel采取的多线程读socket, 顺序写文件, 
Axel这样设计有两个好处: 1) 不用线程锁/解锁; 2) 不用分块儿, 断点续传的时候文件
偏移量就是顺序写文件的暂停位置, 由于浅尝Goroutines, 并行写块儿文件, 并行修改
文件偏移量, 在没有使用任何类似线程概念中的锁/信号/条件, 居然正确写块儿文件! 
这是GoLang的强大? 还是我测试力度还不够呢^_^

GoAxel处理HTTP协议使用的是socket发HTTP HEADER请求, 也尝试过@tuxcanfly的GoDown例子 
https://github.com/tuxcanfly/godown
直接使用GoLang的net/http包, 但是ioutil.ReadAll、io.LimitReader在处理大文件时会
严重阻塞 https://groups.google.com/forum/#!topic/golang-nuts/sAwDldpkMGQ 
所以还是重复造了http轮子, 循环读取socket数据, 没有一点儿阻塞, 哈哈~~~

GoAxel处理FTP协议使用的是@smallfish童鞋的ftp.go包 https://github.com/smallfish/ftp.go 
在此基础上, 添加了下载文件功能 
https://github.com/xiangzhai/goftp/commit/7262e6f47ed02345eab84fbe3cf2ab6b147bb1a7

GoAxel命令行下的进度条使用的是@cheggaaa的pb包 https://github.com/cheggaaa/pb 

GoAxel未来要做的请参看TODO 
