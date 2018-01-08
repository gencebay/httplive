The **HttpLive** is a tool for **API designers, Proxy, mobile and web application developers** to develop and test their applications faster without being dependent on any server or backend applications.

![](https://github.com/gencebay/httplive/blob/master/httplive-ui.png)

HttpLive has a built-in user interface. Therefore, you can do all the configurations you need on this UI, and with dynamic URL (Routing) definitions you can specify your own JSON return types for your applications.

You can share the key-value database (**httplive.db**) with your teammates, easily back up or store it in the any cloud storage.

Load balancing, Round-robin operations can be operated easily with multi-port mode.

With the support of HttpLive you; we can make it more useful without compromising on simple usage and increase the productivity of our development and testing environments.

### Installation

    go get github.com/gencebay/httplive

With this command you can add the **httplive** application to the path you specify in the Go environment. This way you can run the application from the command line.

Make sure your PATH includes the $GOPATH/bin directory so your commands can be easily used with help (-h) flag:

    httplive -h

### Arguments

    --dbpath, -d

Fullpath of the httplive.db with forward slash.

    --ports, -p

Hosting ports can be array comma separated string <5003,5004> to host multiple endpoint. First value of the array is the default port.

HttpLive creates a key-value database for the URLs you define. Here the port value is used as a **bucket name** for the keys. When working with a single port, the data is associated with this port as a keys. When you specify multiple ports, the first value of the array is selected as the default port, and the other ports use the data specified for the default port.

For httplive application running with port 5003:

    GET/api/guideline/mobiletoken

this key will be stored in the **bucket 5003**. Therefor if you running app as single port with 5004 you can not access the keys of 5003 port. You can use multi-port host to overcome this situation.

### Compiling the UI into the Go binary

    go get github.com/jteeuwen/go-bindata/...
    go-bindata -pkg "lib" -o "./lib/bindata.go" public/...

### Todo

Tests

CI Build Integration.

Simple console to display the information of the incoming request under the UI editor. (WebSocket)

Upload a database file from the web interface.

[Watch the video](https://youtu.be/AG5_llcBogk)
