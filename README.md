goaxel
======
Goroutines Download Accelerator

Originally comes from [this repository](https://github.com/xiangzhai/goaxel) created by @xiangzhai

But i add features like user-defined headers, mozilla cookies.txt support

install 
=======
```
go get github.com/kumakichi/goaxel 
go get github.com/cheggaaa/pb
go build
```

usage 
=====
* goaxel -header="Referer:http://ref.test.com;Accept-Encoding:gzip, deflate" -d -o out.aac -U="Mozilla/5.0 (X11; Linux i686; rv:32.0) Gecko/20100101 Firefox/32.0" -n=2 -load-cookies="./cookies.txt" -p /dev/shm/temp http://file.test.com/test.zip?i=0&j=1
* goaxel -h

screenshot 
==========
![ScreenShot](https://raw.github.com/kumakichi/goaxel/master/doc/screenshot.png)