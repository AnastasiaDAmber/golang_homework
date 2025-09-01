package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	head *ListItem
	tail *ListItem
	len  int
}

func NewList() List {
	return &list{}
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.head
}

func (l *list) Back() *ListItem {
	return l.tail
}

func (l *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{Value: v, Next: l.head}
	if l.head != nil {
		l.head.Prev = item
	} else {
		// список был пуст
		l.tail = item
	}
	l.head = item
	l.len++
	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := &ListItem{Value: v, Prev: l.tail}
	if l.tail != nil {
		l.tail.Next = item
	} else {
		// список был пуст
		l.head = item
	}
	l.tail = item
	l.len++
	return item
}

func (l *list) Remove(i *ListItem) {
	if i == nil || l.len == 0 {
		return
	}

	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		// удаление первого элемента списка
		l.head = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		// удаление последнего элемента списка
		l.tail = i.Prev
	}

	i.Next = nil
	i.Prev = nil
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || l.head == i {
		return
	}
	// вырез i
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		// i был хвостом
		l.tail = i.Prev
	}
	// перемещение в начало списка
	i.Prev = nil
	i.Next = l.head
	if l.head != nil {
		l.head.Prev = i
	}
	l.head = i
	if l.tail == nil { // если список был пуст (не уверена, что нужна данная проверка)
		l.tail = i
	}
}
