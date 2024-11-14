package queue

import (
	"errors"
	"sync"
)

type (
	Queue struct {
		start, end *queueItem
		length     int
		mutex      sync.Mutex
	}
	queueItem struct {
		value any
		next  *queueItem
	}
)

// Create a new queue
func New() *Queue {
	return &Queue{
		start:  nil,
		end:    nil,
		length: 0,
	}
}

// Take the next item off the front of the queue
func (q *Queue) Read() any {
	if q.length == 0 {
		return nil
	}
	q.mutex.Lock()

	n := q.start
	if q.length == 1 {
		q.start = nil
		q.end = nil
	} else {
		q.start = q.start.next
	}
	q.length--

	q.mutex.Unlock()
	return n.value
}

// Put an item on the end of a queue
func (q *Queue) Put(value any) {
	q.mutex.Lock()
	n := &queueItem{value, nil}
	if q.length == 0 {
		q.start = n
		q.end = n
	} else {
		q.end.next = n
		q.end = n
	}
	q.length++
	q.mutex.Unlock()
}

// Return the number of items in the queue
func (q *Queue) Len() int {
	return q.length
}

// Return the first item in the queue without removing it
func (q *Queue) Peek() any {
	var result any = nil
	q.mutex.Lock()
	if q.length > 0 {
		result = q.start.value
	}
	q.mutex.Unlock()
	return result
}

// Type for function that using for filtering queue items
type FilterFunc func(value any) bool

// Search element in queue by filter
func (q *Queue) PeekFor(filter FilterFunc) ([]any, error) {
	values := make([]any, 0)
	q.mutex.Lock()

	if q.length > 0 {
		node := q.start
		for {
			if filter(node.value) {
				values = append(values, node.value)
			}
			if node.next == nil {
				break
			}
			node = node.next
		}
	}

	q.mutex.Unlock()
	if len(values) > 0 {
		return values, nil
	}
	return nil, errors.New("values not found")
}

// Take items which filtered
func (q *Queue) ReadFor(filter FilterFunc) ([]any, error) {
	values := make([]any, 0)
	q.mutex.Lock()

	if q.length > 0 {
		curNode := q.start
		q.start = nil
		q.end = nil
		for {
			if filter(curNode.value) {
				values = append(values, curNode.value)
			} else {
				if q.start == nil {
					q.start = curNode
				}
				q.end = curNode
			}

			if curNode.next == nil {
				break
			}
			curNode = curNode.next
		}
		q.length = q.length - len(values)
	}

	q.mutex.Unlock()
	if len(values) > 0 {
		return values, nil
	}
	return nil, errors.New("values not found")
}

// Clear queue
func (q *Queue) Clear() {
	q.mutex.Lock()
	q.start = nil
	q.end = nil
	q.length = 0
	q.mutex.Unlock()
}
