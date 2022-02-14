package loadBalance

import (
	"fmt"
	"testing"
)

func TestConsistent(t *testing.T) {

	hashConsistent := NewConsistent(40)
	hashConsistent.Add("127.0.0.1")
	hashConsistent.Add("192.168.1.11","192.168.0.109")

	// 测试增加虚拟节点后的节点个数
	want := 120
	got := len(hashConsistent.sortedHash)
	isSorted := true

	// 测试哈希环的单调递增性
	for i := 0; i < len(hashConsistent.sortedHash); i++ {
		if i == 0 {
			continue
		}
		if hashConsistent.sortedHash[i] < hashConsistent.sortedHash[i-1] {
			isSorted = false
			break
		}
	}

	// 测试新的请求发来后是否能正常分配服务器
	getTargetIP := "-1"
	getTargetIP, _ = hashConsistent.Get("shirt")
	fmt.Println(getTargetIP)

	if got != want || !isSorted || getTargetIP == "-1" {
		t.Errorf("expected:%v, got%v, isSorted:%v", want, got, isSorted)
	}
}
