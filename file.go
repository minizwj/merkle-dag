package merkledag

const (
	AddressLength = 20
	AddressSize   = 20
	BucketSize    = 3
	NumNodes      = 100
	NumPeers      = 100
	NumBuckets    = 5
	FILE          = iota
	DIR
)

type Node interface {
	Size() uint64
	Name() string
	Type() int
}

type File interface {
	Node

	Bytes() []byte
}

type Dir interface {
	Node

	It() DirIterator
}

type DirIterator interface {
	Next() bool

	Node() Node
}

// ExampleNode 是 Node 接口的一个示例实现
type ExampleNode struct {
	ID   string
	size uint64
}

type Peer struct {
	ID   string
	Size uint64
	DHT  DHT
}

type DHT struct {
	Buckets []Bucket
}

// Bucket 表示 Kademlia DHT 中的一个桶
type Bucket struct {
	Peers []Peer
}

// KBucket 表示 Kademlia DHT 中的 K_Bucket 算法
type KBucket struct {
	Buckets []Bucket
}
