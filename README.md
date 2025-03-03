# Go-BlockChain
BlockChain platform is built by Golang



# BUG:

## Local_Transport

we weil get mesage just  like "REMOTE_1 receiving GetStatus message from REMOTE_1", but the function named SendMessage in localtransport.go cannot send message to ownself.

## TCP_Transport

we will get error about read error, when we run code more than 20 seconds.



:3000服务器的 peerMap为： map[127.0.0.1:4000:0xc0000084f8 127.0.0.1:64429:0xc000008d50 127.0.0.1:64423:0xc00020e000]

```GO
priv := crypto.GeneratePrivateKey()
localNode := makeServer("LOCAL_NDOE", &priv, ":3000", []string{":4000"})
go localNode.Start()

remoteNodeA := makeServer("REMOTE_NODE_A", nil, ":4000", []string{":3000"})
go remoteNodeA.Start()

time.Sleep(7 * time.Second)
lateNode := makeServer("LATE_NODE", nil, ":5000", []string{":3000"})
go lateNode.Start()
```

Map中127.0.0.1:4000是:3000主动连接。其余俩个地址的端口都是主动连接:3000，:3000端口接收到的俩个源地址端口都是操作系统为其随机分配的，为127.0.0.1:64429（:5000的）和127.0.0.1:64423（:4000）



当使用 net.Dial("tcp", addr) 连接到服务器时，conn.LocalAddr() 返回的地址是 本地机器的 IP 地址和一个随机的本地端口号。这个随机端口号是由操作系统分配的，用于标识本地机器的出站连接。

​      为什么是随机端口？因为 TCP 协议是可靠的，所以需要确保两端建立连接后，两端的端口号是相同的。但是，如果本地机器的 IP 地址发生变化，或者本地机器的防火墙设置不当，导致无法建立连接，那么随机端口号就派上了用场。

服务器监听在 :3000，客户端连接到 :4000。 如果服务器和客户端在同一台机器上，连接的四元组可能是：

​      // 源 IP: 127.0.0.1

​      // 源端口: 56752（客户端随机端口）

​      // 目标 IP: 127.0.0.1

​      // 目标端口: 4000
