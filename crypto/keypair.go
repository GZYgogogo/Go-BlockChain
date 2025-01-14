package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"projectx/types"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

// 随机生成私钥
func GeneratePrivateKey() PrivateKey {
	//生成一个椭圆曲线数字签名算法（ECDSA）的私钥，使用的曲线是P256
	//rand.Reader是随机数生成器,确保每次生成的密钥都是唯一且不可预测的。
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return PrivateKey{
		key: key,
	}
}

func (k PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, k.key, data)
	if err != nil {
		return nil, err
	}
	return &Signature{
		R: r,
		S: s,
	}, nil
}

// 通过私钥获取私钥
//
//	func (k PrivateKey) PublicKey() PublicKey {
//		return PublicKey{
//			Key: &k.key.PublicKey,
//		}
//	}
func (k PrivateKey) PublicKey() PublicKey {
	return elliptic.MarshalCompressed(k.key.PublicKey, k.key.PublicKey.X, k.key.PublicKey.Y)
}

// type PublicKey struct {
// 	Key *ecdsa.PublicKey
// }

// func (k PublicKey) ToSlice() []byte {
// 	return elliptic.MarshalCompressed(k.Key, k.Key.X, k.Key.Y)
// }

// func (k PublicKey) Address() types.Address {
// 	h := sha256.Sum256(k.ToSlice())
// 	return types.AddressFromBytes(h[len(h)-20:])
// }

type PublicKey []byte

func (k PublicKey) String() string {
	return hex.EncodeToString(k)
}

func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k)
	return types.AddressFromBytes(h[len(h)-20:])
}

type Signature struct {
	S, R *big.Int
}

// func (s Signature) Verify(pubKey PublicKey, data []byte) bool {
// 	return ecdsa.Verify(pubKey.Key, data, s.R, s.S)
// }

func (sig Signature) Verify(pubKey PublicKey, data []byte) bool {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pubKey)
	key := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	return ecdsa.Verify(key, data, sig.R, sig.S)
}
