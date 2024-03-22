package merkledag

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

func (peer *Peer) SetValue(key, value []byte) bool {
	hash := sha1.Sum(value)
	if string(key) != string(hash[:]) {
		return false
	}

	if peer.DHT.ContainsKeyValue(key) {
		return true
	}

	bucketIndex := peer.DHT.GetBucketIndex(key, peer.ID)

	closestNodes := peer.DHT.SelectClosestNodes(bucketIndex, key, 2)

	for _, node := range closestNodes {
		node.SetValue(key, value)
	}

	return true
}

func (peer *Peer) GetValue(key []byte) []byte {
	if value, ok := peer.DHT.GetKeyValue(key); ok {
		return value
	}

	bucketIndex := peer.DHT.GetBucketIndex(key, peer.ID)

	closestNodes := peer.DHT.SelectClosestNodes(bucketIndex, key, 2)

	for _, node := range closestNodes {
		result := node.GetValue(key)
		if result != nil {
			return result
		}
	}

	return nil
}

func (dht *DHT) ContainsKeyValue(key []byte) bool {
	bucketIndex := dht.GetBucketIndex(key, "")
	bucket := dht.Buckets[bucketIndex]

	for _, peer := range bucket.Peers {
		if peer.DHT.ContainsKeyValue(key) {
			return true
		}
	}
	return false
}

func (dht *DHT) GetKeyValue(key []byte) ([]byte, bool) {
	bucketIndex := dht.GetBucketIndex(key, "")
	bucket := dht.Buckets[bucketIndex]

	for _, peer := range bucket.Peers {
		value, exists := peer.DHT.GetKeyValue(key)
		if exists {
			return value, true
		}
	}
	return nil, false
}

func (dht *DHT) GetBucketIndex(key []byte, nodeID string) int {
	distance := calculateDistance(key, nodeID)
	bucketIndex := logDistance(distance, len(dht.Buckets))
	return bucketIndex
}

func calculateDistance(key []byte, nodeID string) []byte {
	distance := make([]byte, len(key))
	for i := 0; i < len(key); i++ {
		distance[i] = key[i] ^ nodeID[i]
	}
	return distance
}

func logDistance(distance []byte, bucketCount int) int {
	bucketIndex := -1
	for i := 0; i < len(distance); i++ {
		if distance[i] != 0 {
			bucketIndex = i*8 + leadingZerosCount(distance[i])
			break
		}
	}
	if bucketIndex == -1 {
		bucketIndex = len(distance) * 8
	}
	if bucketIndex >= bucketCount {
		bucketIndex = bucketCount - 1
	}
	return bucketIndex
}

func leadingZerosCount(b byte) int {
	count := 0
	for i := 7; i >= 0; i-- {
		if b&(1<<i) != 0 {
			break
		}
		count++
	}
	return count
}

func (dht *DHT) SelectClosestNodes(bucketIndex int, key []byte, count int) []Peer {
	// 在指定桶中选择与指定键最接近的节点
	bucket := dht.Buckets[bucketIndex]
	peerCount := len(bucket.Peers)

	if peerCount <= count {
		return bucket.Peers
	}

	// 按与指定键的距离进行排序
	sort.Slice(bucket.Peers, func(i, j int) bool {
		distanceI := calculateDistance(key, bucket.Peers[i].ID)
		distanceJ := calculateDistance(key, bucket.Peers[j].ID)

		for k := 0; k < len(distanceI); k++ {
			if distanceI[k] < distanceJ[k] {
				return true
			} else if distanceI[k] > distanceJ[k] {
				return false
			}
		}

		return false
	})

	return bucket.Peers[:count]
}

func generateRandomID() string {
	id := make([]byte, AddressSize)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(id)
}

func generateRandomKey() []byte {
	key := make([]byte, AddressSize)
	rand.Read(key)
	return key
}

func generateRandomValue() []byte {
	value := make([]byte, 16)
	rand.Read(value)
	return value
}
func main() {
	rand.Seed(time.Now().UnixNano())

	nodes := make([]Peer, NumNodes)
	for i := 0; i < NumNodes; i++ {
		// 初始化每个节点的ID和DHT
		node := Peer{
			ID: generateRandomID(),
			DHT: DHT{
				Buckets: make([]Bucket, NumBuckets),
			},
		}
		nodes[i] = node
	}

	keys := make([][]byte, 200)
	values := make([][]byte, 200)
	for i := 0; i < 200; i++ {
		key := generateRandomKey()
		value := generateRandomValue()
		keys[i] = key
		values[i] = value
	}

	for i := 0; i < 200; i++ {
		nodeIndex := rand.Intn(NumNodes)
		node := &nodes[nodeIndex]
		node.SetValue(keys[i], values[i])
	}

	for i := 0; i < 100; i++ {
		keyIndex := rand.Intn(200)
		key := keys[keyIndex]
		nodeIndex := rand.Intn(NumNodes)
		node := &nodes[nodeIndex]
		result := node.GetValue(key)
		if result != nil {
			fmt.Println("找到键:", string(key))
			fmt.Println("值:", string(result))
		}
	}
}
