package skiplist

import (
	"testing"
)

func TestCreate(t *testing.T) {
	list := Create()
	list.Insert(100, "a")
	list.Insert(200, "b")
	list.Insert(300, "c")
	list.Insert(400, "e")
	list.Insert(500, "f")
	list.Insert(600, "g")
	list.Insert(700, "h")
	t.Log(list.GetRank(200, "b"))
	e := list.GetElementByRank(5)
	if e != nil {
		t.Log(e.Score)
	}
}
