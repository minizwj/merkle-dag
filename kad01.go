package merkledag

import (
	"fmt"
	"math/rand"
	"strings"
)

func (en *ExampleNode) Size() uint64 {
	return en.size
}

func (en *ExampleNode) Name() string {
	return en.ID
}

func (en *ExampleNode) Type() int {
	return 1
}

func (kb *KBucket) InsertNode(node Node) {
	bucketIndex := kb.getBucketIndex(node.Name())

	if len(kb.Buckets[bucketIndex].Peers) >= BucketSize {
		if kb.coversBucket(bucketIndex, node.Name()) {
			kb.splitBucket(bucketIndex)
			bucketIndex = kb.getBucketIndex(node.Name())
		} else {
			return
		}
	}

	kb.Buckets[bucketIndex].Peers = append(kb.Buckets[bucketIndex].Peers, Peer{ID: node.Name(), Size: node.Size()})
}

func (kb *KBucket) PrintBucketContents() {
	for i, bucket := range kb.Buckets {
		peerNames := make([]string, len(bucket.Peers))
		for j, peer := range bucket.Peers {
			peerNames[j] = peer.ID
		}
		fmt.Printf("Bucket %d: %s\n", i, strings.Join(peerNames, ", "))
	}
}

func (kb *KBucket) getBucketIndex(nodeName string) int {
	prefixLength := commonPrefixLength(kb.Buckets[0].Peers[0].ID, nodeName)
	return prefixLength
}

func (kb *KBucket) coversBucket(bucketIndex int, nodeName string) bool {
	prefixLength := commonPrefixLength(kb.Buckets[bucketIndex].Peers[0].ID, nodeName)
	return prefixLength > bucketIndex
}

func (kb *KBucket) splitBucket(bucketIndex int) {
	oldBucket := kb.Buckets[bucketIndex]
	newBucket := Bucket{}

	randomPeerIndex := rand.Intn(len(oldBucket.Peers))
	randomPeer := oldBucket.Peers[randomPeerIndex]

	oldBucket.Peers = append(oldBucket.Peers[:randomPeerIndex], oldBucket.Peers[randomPeerIndex+1:]...)

	for _, peer := range oldBucket.Peers {
		if kb.getBucketIndex(peer.ID) == bucketIndex {
			newBucket.Peers = append(newBucket.Peers, peer)
		}
	}

	newBucket.Peers = append(newBucket.Peers, randomPeer)

	kb.Buckets[bucketIndex] = oldBucket
	kb.Buckets = append(kb.Buckets, Bucket{})
	copy(kb.Buckets[bucketIndex+1:], kb.Buckets[bucketIndex:])
	kb.Buckets[bucketIndex] = newBucket
}

func commonPrefixLength(str1, str2 string) int {
	length := 0
	for i := 0; i < len(str1) && i < len(str2); i++ {
		if str1[i] == str2[i] {
			length++
		} else {
			break
		}
	}
	return length
}

func main() {
	// 创建一个示例的 KBucket
	kb := KBucket{
		Buckets: make([]Bucket, AddressLength),
	}

	// 创建一些示例节点
	nodes := []Node{
		&ExampleNode{ID: "node1", size: 10},
		&ExampleNode{ID: "node2", size: 15},
		&ExampleNode{ID: "node3", size: 8},
		&ExampleNode{ID: "node4", size: 12},
		&ExampleNode{ID: "node5", size: 6},
	}

	// 将节点插入 KBucket
	for _, node := range nodes {
		kb.InsertNode(node)
	}

	// 打印每个桶中的节点名称
	kb.PrintBucketContents()
}
