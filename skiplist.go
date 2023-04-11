package skiplist

import (
	"math/rand"
	"strings"
)

const (
	MaxLevel         = 32
	P        float64 = 0.25
)

type Level struct {
	forward *Node
	span    uint64
}

type Node struct {
	Obj   string
	Score float64

	backward *Node
	level    []Level
}

type List struct {
	Length uint64

	header, tail *Node
	level        uint32
}

func Create() *List {
	l := &List{
		Length: 0,
		level:  1,
	}
	l.header = &Node{
		level: make([]Level, MaxLevel),
	}

	return l
}

func randomLevel() uint32 {
	var level uint32 = 1
	for rand.Float64() < P && level < MaxLevel {
		level += 1
	}

	return level
}

func (l *List) Insert(score float64, obj string) *Node {
	update := make([]*Node, MaxLevel)
	rank := make([]uint64, MaxLevel)

	// 在各个层查找节点的插入位置
	x := l.header
	for i := l.level; i > 0; i-- {
		j := i - 1
		// 如果 i 不是 zsl->level-1 层
		// 那么 i 层的起始 rank 值为 i+1 层的 rank 值
		// 各个层的 rank 值一层层累积
		// 最终 rank[0] 的值加一就是新节点的前置节点的排位
		// rank[0] 会在后面成为计算 span 值和 rank 值的基础
		if j < l.level-1 {
			rank[j] = rank[j+1]
		}
		for x.level[j].forward != nil &&
			(x.level[j].forward.Score < score ||
				(x.level[j].forward.Score == score && strings.Compare(x.level[j].forward.Obj, obj) < 0)) {
			// 记录沿途跨越了多少个节点
			rank[j] += x.level[j].span
			// 移动至下一指针
			x = x.level[j].forward
		}
		// 记录将要和新节点相连接的节点
		update[j] = x
	}

	// 获取一个随机值作为新节点的层数
	level := randomLevel()

	// 如果新节点的层数比表中其他节点的层数都要大
	// 那么初始化表头节点中未使用的层，并将它们记录到 update 数组中
	// 将来也指向新节点
	if level > l.level {
		for i := l.level; i < level; i++ {
			rank[i] = 0
			update[i] = l.header
			l.header.level[i].span = l.Length
		}

		// 更新表中节点最大层数
		l.level = level
	}

	// 创建新节点
	x = &Node{
		Obj:   obj,
		Score: score,
		level: make([]Level, level),
	}

	// 将前面记录的指针指向新节点，并做相应的设置
	for i := 0; i < int(level); i++ {
		// 设置新节点的 forward 指针
		x.level[i].forward = update[i].level[i].forward
		// 将沿途记录的各个节点的 forward 指针指向新节点
		update[i].level[i].forward = x
		// 计算新节点跨越的节点数量
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		// 更新新节点插入之后，沿途节点的 span 值
		// 其中的 +1 计算的是新节点
		update[i].level[i].span = rank[0] - rank[i] + 1
	}

	// 未接触的节点的 span 值也需要增一，这些节点直接从表头指向新节点
	for i := level; i < l.level; i++ {
		update[i].level[i].span += 1
	}

	// 设置新节点的后退指针
	if update[0] != l.header {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		l.tail = x
	}

	// 跳跃表的节点计数增一
	l.Length += 1

	return x
}

func (l *List) deleteNode(x *Node, update []*Node) {
	// 更新所有和被删除节点 x 有关的节点的指针，解除它们之间的关系
	for i := 0; i < int(l.level); i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].forward = x.level[i].forward
			update[i].level[i].span += x.level[i].span - 1
		} else {
			update[i].level[i].span -= 1
		}
	}

	// 更新被删除节点 x 的前进和后退指针
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		l.tail = x.backward
	}

	// 更新跳跃表最大层数（只在被删除节点是跳跃表中最高的节点时才执行）
	for l.level > 1 && l.header.level[l.level-1].forward == nil {
		l.level -= 1
	}

	// 跳跃表节点计数器减一
	l.Length -= 1
}

func (l *List) Delete(score float64, obj string) error {
	update := make([]*Node, MaxLevel)

	// 遍历跳跃表，查找目标节点，并记录所有沿途节点
	x := l.header
	for i := l.level; i > 0; i-- {
		j := i - 1
		for x.level[j].forward != nil &&
			(x.level[j].forward.Score < score ||
				(x.level[j].forward.Score == score && strings.Compare(x.level[j].forward.Obj, obj) < 0)) {
			// 沿着前进指针移动
			x = x.level[j].forward
		}
		// 记录沿途节点
		update[j] = x
	}

	// 检查找到的元素 x ，只有在它的分值和对象都相同时，才将它删除。
	x = x.level[0].forward
	if x != nil && x.Score == score && strings.Compare(x.Obj, obj) == 0 {
		l.deleteNode(x, update)
		return nil
	}

	return ErrNodeNotFound
}

func (l *List) GetRank(score float64, obj string) uint64 {
	var rank uint64

	// 遍历整个跳跃表
	x := l.header
	for i := l.level; i > 0; i-- {
		j := i - 1
		for x.level[j].forward != nil &&
			(x.level[j].forward.Score < score ||
				(x.level[j].forward.Score == score && strings.Compare(x.level[j].forward.Obj, obj) <= 0)) {
			// 累计跨越的节点数
			rank += x.level[j].span
			// 沿着前进指针移动
			x = x.level[j].forward
		}

		if x != nil && strings.Compare(x.Obj, obj) == 0 {
			return rank
		}
	}

	// 没找到
	return 0
}

func (l *List) GetElementByRank(rank uint64) *Node {
	if rank > l.Length && rank == 0 {
		return nil
	}

	x := l.header
	for i := l.level; i > 0; i-- {
		j := i - 1
		for x.level[j].forward != nil && rank >= x.level[j].span {
			// 减去以及越过的节点数
			rank -= x.level[j].span
			x = x.level[j].forward
		}
		if rank == 0 {
			return x
		}
	}

	return nil
}
