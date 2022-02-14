package common

import (
	"testing"
)

func TestConsistent_Add(t *testing.T) {
	hostArray := []string{
		"127.0.0.1",
		"192.168.1.11",
	}

	hashConsistent := NewConsistent()
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}

	want := 40
	got := len(hashConsistent.sortedHash)
	isSorted := true
	for i := 0; i < len(hashConsistent.sortedHash); i++ {
		if i == 0 {
			continue
		}
		if hashConsistent.sortedHash[i] < hashConsistent.sortedHash[i-1] {
			isSorted = false
			break
		}
	}

	if got != want || !isSorted {
		t.Errorf("expected:%v, got%v, isSorted:%v", want, got, isSorted)
	}
}
