package common

import (
	"github.com/pkg/errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// 声明新切片类型
type units []uint32

func (x units) Len() int {
	return len(x)
}

func (x units) Less(i, j int) bool {
	return x[i] < x[j]
}

func (x units) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

var errEmpty = errors.New("Hash 环没有数据")

// 创建结构体保存一致性
type Consistent struct {
	circle       map[uint32]string // hash环，key为hash值，值为节点的信息
	sortedHash   units             // 已经排序的hash切片
	VirtualNode  int               // 虚拟节点个数，用来增加hash的平衡性
	sync.RWMutex                   // 读写锁
}

// 创建一致性hash算法结构体，设置默认节点数量
func NewConsistent() *Consistent {
	return &Consistent{
		circle:      make(map[uint32]string),
		VirtualNode: 20,
	}
}

// 自动生成key值
func (c *Consistent) generateKey(element string, index int) string {
	return element + strconv.Itoa(index)
}

// 获取hash位置
func (c *Consistent) hashKey(key string) uint32 {
	if len(key) < 64 {
		scratch := make([]byte, 64)
		copy(scratch, key)
		// 使用IEEE 多项式返回数据的CRC-32位校验和
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

// 更新排序，方便查找
func (c *Consistent) updateSortedHashs() {
	hashes := c.sortedHash[:0]

	// 判断切片容量，是否过大，如果过大则重置
	if cap(c.sortedHash)/(c.VirtualNode*4) > len(c.circle) {
		hashes = nil
	}

	for k := range c.circle {
		hashes = append(hashes, k)
	}

	// 对所有节点hash值进行排序，方便之后二分查找
	sort.Sort(hashes)
	c.sortedHash = hashes
}

// 向hash环添加节点
func (c *Consistent) Add(element string) {
	// 加锁
	c.Lock()
	defer c.Unlock()
	c.add(element)
}

// 添加节点
func (c *Consistent) add(element string) {
	// 循环虚拟节点，设置副本
	for i := 0; i < c.VirtualNode; i++ {
		c.circle[c.hashKey(c.generateKey(element, i))] = element
	}

	// 更新排序
	c.updateSortedHashs()
}

func (c *Consistent) Remove(element string) {
	// 加锁
	c.Lock()
	defer c.Unlock()
	c.remove(element)
}

// 删除一个节点
func (c *Consistent) remove(element string) {
	for i := 0; i < c.VirtualNode; i++ {
		delete(c.circle, c.hashKey(c.generateKey(element, i)))
		c.updateSortedHashs()
	}
}

// 顺时针查找对应的节点
func (c *Consistent) search(key uint32) int {
	// 查找
	f := func(x int) bool {
		return c.sortedHash[x] > key
	}

	// 二分查找搜索满足第一个大于key的值
	i := sort.Search(len(c.sortedHash), f)

	// 如果查处范围，设置i=0
	if i >= len(c.sortedHash) {
		i = 0
	}
	return i
}

// 根据数据标识获取最近的服务器节点信息
func (c *Consistent) Get(name string) (string, error) {
	c.RLock()
	defer c.RUnlock()
	// 如果为0，则返回错误
	if len(c.circle) == 0 {
		return "", errEmpty
	}

	// 计算hash值
	key := c.hashKey(name)
	i := c.search(key)
	return c.circle[c.sortedHash[i]], nil
}
