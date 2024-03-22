//package merkledag
//
//import "hash"
//
//type Link struct {
//	Name string
//	Hash []byte
//	Size int
//}
//
//type Object struct {
//	Links []Link
//	Data  []byte
//}
//
//func Add(store KVStore, node Node, h hash.Hash) []byte {
//	// TODO 将分片写入到KVStore中，并返回Merkle Root
//	return nil
//}

package merkledag

import (
	"bytes"
	"encoding/gob"
	"hash"
	"math"
)

const (
	Blob        = 256 * 1024
	MaxBlobList = 1024
)

type Link struct {
	Hash string // 子节点的哈希值
	Size int    // 子节点的大小
}

type Object struct {
	Data  []string // 类型（blob、list、tree）
	Links []Link
}

// 序列化对象
func serialize(obj *Object) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func storeBlob(store KVStore, data []byte, h hash.Hash, length int) ([]byte, string) {
	h.Reset()
	h.Write(data)
	hashValue := h.Sum(nil)

	links := []Link{
		{
			Hash: string(hashValue),
			Size: length,
		},
	}

	obj := &Object{
		Data:  []string{"blob"},
		Links: links,
	}

	objBytes, _ := serialize(obj)
	err := store.Put(hashValue, objBytes) // 存储序列化后的数据
	if err != nil {
		return nil, ""
	}
	return hashValue, "blob"
}

func storeBlobList(store KVStore, links []Link, h hash.Hash) ([]byte, string) {
	data := make([]string, len(links))
	for i := range data {
		data[i] = "blob"
	}

	obj := &Object{
		Data:  data,
		Links: links,
	}

	objBytes, _ := serialize(obj)

	h.Reset() // 重置哈希状态
	h.Write(objBytes)
	listHash := h.Sum(nil)

	store.Put(listHash, objBytes) // 存储序列化后的数据
	return listHash, "list"
}

func storeFile(store KVStore, node Node, h hash.Hash) ([]byte, string) {
	file := node.(File)
	fileContent := file.Bytes()

	// 计算需要分割成的块数
	numBlobs := int(math.Ceil(float64(len(fileContent)) / float64(Blob)))
	blobLinks := make([]Link, numBlobs)
	blobData := make([]string, numBlobs)

	if numBlobs > 1 && numBlobs <= MaxBlobList {
		// 当块数大于1且小于等于MaxBlobList时，将块存储到列表中
		for i := 0; i < numBlobs; i++ {
			start := i * Blob
			end := int(math.Min(float64(start+Blob), float64(len(fileContent))))
			blobChunk := fileContent[start:end]

			blobHash, _ := storeBlob(store, blobChunk, h, end-start)

			link := Link{
				Hash: string(blobHash),
				Size: end - start,
			}
			blobLinks[i] = link
			blobData[i] = "blob"
		}

		return storeBlobList(store, blobLinks, h)
	} else {
		// 当块数大于MaxBlobList时，将块存储到多个列表中
		numLists := int(math.Ceil(float64(numBlobs) / float64(MaxBlobList)))
		listLinks := make([]Link, numLists)
		listData := make([]string, numLists)

		for i := 0; i < numLists; i++ {
			start := i * MaxBlobList
			end := int(math.Min(float64(start+MaxBlobList), float64(numBlobs)))
			listLinksChunk := blobLinks[start:end]
			listChunkSize := 0

			for j := start; j < end; j++ {
				listChunkSize += blobLinks[j].Size
			}

			listHash, _ := storeBlobList(store, listLinksChunk, h)

			link := Link{
				Hash: string(listHash),
				Size: len(listLinksChunk),
			}
			listLinks[i] = link
			listData[i] = "list"
		}

		obj := &Object{
			Data:  listData,
			Links: listLinks,
		}

		objBytes, _ := serialize(obj)
		h.Reset() // 重置哈希状态
		h.Write(objBytes)
		listHash := h.Sum(nil)
		store.Put(listHash, objBytes)
		return storeBlobList(store, listLinks, h)
	}
}

// 将目录节点存储到数据库中
func storeDir(store KVStore, node Node, h hash.Hash) []byte {
	dir := node.(Dir)
	childLinks := make([]Link, 0, dir.Size())
	childData := make([]string, 0, dir.Size())

	iterator := dir.It()

	for iterator.Next() {
		child := iterator.Node()
		if child.Type() == FILE {
			childHash, childType := storeFile(store, child, h)
			link := Link{
				Hash: string(childHash),
				Size: int(child.Size()),
			}
			childLinks = append(childLinks, link)
			childData = append(childData, childType)

		} else if child.Type() == DIR {
			dirHash := storeDir(store, child, h)
			nodeType := "tree"
			if len(childLinks) == 0 {
				nodeType = "blob"
			}
			childLinks = append(childLinks, Link{string(dirHash), int(child.Size())})
			childData = append(childData, nodeType)
		}
	}

	obj := &Object{
		Data:  childData,
		Links: childLinks,
	}

	objBytes, err := serialize(obj)
	if err != nil {
		return nil
	}

	dirHash := h.Sum(objBytes)

	store.Put(dirHash, objBytes)

	return dirHash
}

// Add函数用于生成MerkleDAG
func Add(store KVStore, node Node, h hash.Hash) []byte { //返回roothash和当前节点的类型
	//TODO 将分片写入到KVStore中，并返回Merkle Root
	var rootHash []byte
	switch node.Type() {
	case FILE:
		rootHash, _ = storeFile(store, node, h)
		return rootHash

	case DIR:
		rootHash = storeDir(store, node, h)
		return rootHash
	default:
		return nil

	}
}
