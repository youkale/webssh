package tui

type fixedLengthVec struct {
	len   int
	items []interface{}
	head  int
	tail  int
	size  int
}

func newFixedLengthVec(len int) *fixedLengthVec {
	return &fixedLengthVec{
		len:   len,
		items: make([]interface{}, len),
		head:  0,
		tail:  0,
		size:  0,
	}
}

func (q *fixedLengthVec) Push(item interface{}) {
	if q.size == q.len {
		q.head = (q.head + 1) % q.len
	} else {
		q.size++
	}
	q.items[q.tail] = item
	q.tail = (q.tail + 1) % q.len
}

func (q *fixedLengthVec) Pop() (interface{}, bool) {
	if q.size == 0 {
		return "", false
	}
	item := q.items[q.head]
	q.head = (q.head + 1) % q.len
	q.size--
	return item, true
}

func (q *fixedLengthVec) Len() int {
	return q.size
}

func (q *fixedLengthVec) Items() []interface{} {
	if q.size == 0 {
		return nil
	}
	result := make([]interface{}, q.size)
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.len
		result[i] = q.items[idx]
	}
	return result
}
