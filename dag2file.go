package merkledag

import (
	"bytes"
	"encoding/gob"
	"strings"
)

//func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
//	// 根据hash和path， 返回对应的文件, hash对应的类型是tree
//	return nil
//}
//

func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
	rootObjBytes, _ := store.Get(hash)
	rootObj, _ := deserialize(rootObjBytes)
	currentNode := rootObj
	pathComponents := splitPath(path)

	for _, component := range pathComponents {
		if currentNode.Data[0] != "tree" {
			return nil
		}

		childHash := findChildHash(currentNode, component)
		if childHash == nil {
			return nil
		}
		childObjBytes, _ := store.Get(childHash)
		childObj, _ := deserialize(childObjBytes)

		currentNode = childObj
	}

	if currentNode.Data[0] != "blob" {
		return nil
	}

	fileBytes, _ := store.Get([]byte(currentNode.Links[0].Hash))
	return fileBytes
}

// 拆分路径
func splitPath(path string) []string {
	return strings.Split(path, "/")
}

// 反序列化object
func deserialize(data []byte) (*Object, error) {
	var obj Object
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

// 根据子节点名称查找对应的哈希值
func findChildHash(node *Object, name string) []byte {
	for _, link := range node.Links {
		if link.Name == name {
			return []byte(link.Hash)
		}
	}
	return nil
}
