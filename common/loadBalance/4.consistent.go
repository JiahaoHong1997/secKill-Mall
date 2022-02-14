package loadBalance

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type units []uint32

func (u units) Len() int {
	return len(u)
}

func (u units) Less(i, j int) bool {
	return u[i] < u[j]
}

func (u units) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

type Consistent struct {
	circle      map[uint32]string
	sortedHash  units
	VirtualNode int
	l           sync.RWMutex
}

func NewConsistent(k int) *Consistent {
	return &Consistent{
		circle: make(map[uint32]string),
		VirtualNode: k,
	}
}

func (c *Consistent) generateKey(element string, index int) string {
	return element + strconv.Itoa(index)
}


func (c *Consistent) hashKey(key string) uint32 {
	if len(key) < 64 {
		t := make([]byte, 64)
		copy(t, key)
		return crc32.ChecksumIEEE(t[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *Consistent) updateSortedHashes() {
	hashes := c.sortedHash[:0]

	for k := range c.circle {
		hashes = append(hashes, k)
	}

	sort.Sort(hashes)
	c.sortedHash = hashes
}

func (c *Consistent) add(element ...string) {

	for k:=0; k<len(element); k++ {
		x := element[k]
		for i:=0; i<c.VirtualNode; i++ {
			c.circle[c.hashKey(c.generateKey(x, i))] = x
		}
	}


	c.updateSortedHashes()
}

func (c *Consistent) Add(element ...string) error {
	c.l.Lock()
	defer c.l.Unlock()

	for i:=0; i<len(element); i++ {
		c.add(element[i])
	}
	return nil
}

func (c *Consistent) remove(element string) {
	for i:=0; i<c.VirtualNode; i++ {
		delete(c.circle, c.hashKey(c.generateKey(element, i)))
	}
	c.updateSortedHashes()
}

func (c *Consistent) Remove(element string) {
	c.l.Lock()
	defer c.l.Unlock()
	c.remove((element))
}

func (c *Consistent) get(key uint32) int {
	i := sort.Search(len(c.sortedHash), func(x int) bool {
		return c.sortedHash[x] > key
	})

	if i>= len(c.sortedHash) {
		i = 0
	}
	return i
}

func (c *Consistent) Get(name string) (string, error) {
	c.l.RLock()
	defer c.l.RUnlock()

	if len(c.circle) == 0 {
		return "", errors.New("No datas on hash circle")
	}
	key := c.hashKey(name)
	i := c.get(key)
	return c.circle[c.sortedHash[i]], nil
}
