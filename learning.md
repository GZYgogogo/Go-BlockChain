# makefile

下载好mingw、配置好gcc路径



# Test模块

依赖包："testing"

函数名称以Test开头，TestConnect ( t *testing.T )

go run -v ./...		ex：go.exe test -timeout 30s  projectx/core.go

使用makeFile运行test

test功能同时需要正确和错误俩面



# 传输协议

搭建自己本地的传输层协议，而非使用tcp或者udp协议

本网络中区块链的数据传输都是通过该协议进行的





# GO

## 多精度

big.Int 表示有符号多精度整数

## 泛型接口

在 Go 中，**接口**是一组方法的集合。如果一个类型实现了某个接口要求的所有方法，则该类型自动实现了该接口。

```go
type Hasher[T any] interface {
  Hash(T) types.Hash
}
```

此接口定义了一个方法 `Hash`，它接受一个类型为 `T` 的参数，并返回一个 `types.Hash` 类型的结果。这意味着任何实现该接口的类型都必须提供具体的 `Hash` 方法，该方法能够处理 `T` 类型的输入并返回一个 `types.Hash`

提供一个统一的哈希函数框架，使得不同数据类型的实例都可以通过实现 `Hash` 方法来计算其哈希值。



# 项目：

传输协议是由自己定义的

外部程序调用的api接口由RPC协议封装

