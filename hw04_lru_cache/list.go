package hw04lrucache

type List interface {
	// Len длина списка
	Len() int
	// Front первый элемент списка
	Front() *ListItem
	// Back последний элемент списка
	Back() *ListItem
	// PushFront добавить значение в начало
	PushFront(v interface{}) *ListItem
	// PushBack добавить значение в конец
	PushBack(v interface{}) *ListItem
	// Remove удалить элемент
	Remove(i *ListItem)
	// MoveToFront переместить элемент в начало
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	front *ListItem
	back  *ListItem
	len   int
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	newItem := &ListItem{Value: v}

	if l.front == nil {
		l.front = newItem
		l.back = newItem
	} else {
		newItem.Next = l.front
		l.front.Prev = newItem
		l.front = newItem
	}

	l.len++
	return newItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	newItem := &ListItem{Value: v}

	if l.back == nil {
		l.front = newItem
		l.back = newItem
	} else {
		newItem.Prev = l.back
		l.back.Next = newItem
		l.back = newItem
	}

	l.len++
	return newItem
}

func (l *list) Remove(i *ListItem) {
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.front = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.back = i.Prev
	}

	i.Next = nil
	i.Prev = nil
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == l.front {
		return
	}

	l.Remove(i)

	i.Next = l.front
	i.Prev = nil

	if l.front != nil {
		l.front.Prev = i
	}

	l.front = i

	if l.back == nil {
		l.back = i
	}

	l.len++
}

func NewList() List {
	return new(list)
}
