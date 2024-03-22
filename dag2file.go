//package merkledag
//
//import (
//	"fmt"
//	"io/ioutil"
//)
//
//// Hash to file
//func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
//	// 根据hash和path， 返回对应的文件, hash对应的类型是tree
//	return nil
//}
//

package merkledag

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

// Hash2File 根据哈希值和路径，从 KVStore 中读取相应的文件数据并返回
func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
	// 根据hash和path， 返回对应的文件, hash对应的类型是tree
	// 获取哈希
	hasher := hp.Get()

	// 将哈希值写入哈希函数
	_, err := hasher.Write(hash)
	if err != nil {
		return nil
	}

	// 计算哈希值
	calculatedHash := hasher.Sum(nil)

	// 校验哈希值是否匹配
	if !bytes.Equal(hash, calculatedHash) {
		fmt.Printf("hash mismatch")
		return nil
	}

	// 从 KVStore 中获取与哈希值相关联的数据
	data, err := store.Get(hash)
	if err != nil {
		return nil
	}

	// 数据写入文件
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return nil
	}

	// 文件中读取内容
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	return fileData
}
