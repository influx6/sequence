package sequence

import "errors"

import "sync"

const (
	//MINBUFF states the default minimum buffer size for the write channels
	MINBUFF = 20
)

var (
	//ErrBADValue represents a bad value calculation by the iterator
	ErrBADValue = errors.New("Bad Value!")
	//ErrBADINDEX represents a bad index counter by the iterator
	ErrBADINDEX = errors.New("Bad Index!")
	//ErrENDINDEX represents a reaching of the end of an iterator
	ErrENDINDEX = errors.New("End Index!")
)

//MutFunc is the type of a function whoes argument is a Sequencable
type MutFunc func(f interface{}) interface{}

//ProcFunc is the type of a function giving to a BaseIterator
type ProcFunc func(f Iterable) (interface{}, interface{}, error)

//Iterable defines sequence method rules
type Iterable interface {
	Next() error
	Key() interface{}
	Value() interface{}
	Reset()
	Length() int
	Clone() Iterable
}

//Sequencable defines a sequence method rules
type Sequencable interface {
	Iterator() Iterable
	Parent() Sequencable
	// Value() interface{}
}

//MutateSequencable defines the methods of a sequence to be able to mutate
type MutateSequencable interface {
	Sequencable
	Mutate(MutFunc)
}

//Sizable provides a length data member
type Sizable interface {
	Length() int
}

//SizableSequencable provides sequences with size detail
type SizableSequencable interface {
	Sequencable
	Sizable
}

//MutableSizableSequencable provides sequences with size detail
type MutableSizableSequencable interface {
	MutateSequencable
	Sizable
}

//ListSequencable defines ListSequence method rules
type ListSequencable interface {
	MutableSizableSequencable
	Obj() []interface{}
	Clear() ListSequencable
	Add(...interface{}) ListSequencable
	Delete(...interface{}) ListSequencable
	Get(interface{}) interface{}
	Clone() ListSequencable
	Keys() ListSequencable
	Values() ListSequencable
}

//MapSequencable defines MapSequence method rules
type MapSequencable interface {
	MutableSizableSequencable
	Obj() map[interface{}]interface{}
	Clear() MapSequencable
	Add(...interface{}) MapSequencable
	Delete(...interface{}) MapSequencable
	Get(interface{}) interface{}
	Clone() MapSequencable
	Keys() ListSequencable
	Values() ListSequencable
}

//IterableSequence is the root level of immutable sequence types
type IterableSequence struct {
	*Sequence
	iterator Iterable
}

//MixSequence takes a sequence and returns a sequence based on its iterator
func MixSequence(s Sequencable) Sequencable {
	return NewIterableSequence(s.Iterator())
}

//Iterator returns a new base iterator for the sequence
func (t *IterableSequence) Iterator() Iterable {
	return IdentityIterator(t.iterator)
}

//Parent returns a new base iterator for the sequence
func (t *IterableSequence) Parent() Sequencable {
	return Sequencable(t)
}

//Value returns a the final value of a sequence operation
func (t *IterableSequence) Value() interface{} {
	return t.iterator.Value()
}

//NewIterableSequence returns a new sequence based off an iterable
func NewIterableSequence(f Iterable) *IterableSequence {
	return &IterableSequence{
		NewBaseSequence(0, nil),
		f,
	}
}

//Sequence is the root level structure for all sequence types
type Sequence struct {
	parent Sequencable
	// writer *SeqWriter
	lock *sync.RWMutex
}

//Iterator returns the iterator of the sequence
func (s *Sequence) Iterator() Iterable {
	if s.parent == nil {
		return nil
	}
	return s.parent.Iterator()
}

//Parent returns a sequencable
func (s *Sequence) Parent() Sequencable {
	if s.parent == nil {
		return nil
	}
	return s.parent.Parent()
}

//NewBaseSequence returns a base sequence struct
func NewBaseSequence(buff int, parent Sequencable) *Sequence {
	if buff < MINBUFF {
		buff = MINBUFF
	}

	return &Sequence{
		parent,
		new(sync.RWMutex),
	}
}

//NewListSequence returns a new ListSequence
func NewListSequence(data []interface{}, buff int) *ListSequence {
	if data == nil {
		data = make([]interface{}, 0)
	}

	return &ListSequence{
		NewBaseSequence(buff, nil),
		data,
		buff,
	}
}

//NewMapSequence returns a new MapSequence
func NewMapSequence(data map[interface{}]interface{}, buff int) *MapSequence {
	if data == nil {
		data = make(map[interface{}]interface{})
	}

	return &MapSequence{
		NewBaseSequence(buff, nil),
		data,
		buff,
	}
}

//MapSequence represents a sequence for maps
type MapSequence struct {
	*Sequence
	data   map[interface{}]interface{}
	buffer int
}

//Mutate allows mutation on sequence data
func (l *MapSequence) Mutate(fn MutFunc) {
	l.lock.Lock()
	res, ok := fn(l.data).(map[interface{}]interface{})

	if !ok {
		return
	}

	l.data = res
	l.lock.Unlock()
}

//Iterator returns the sequence data iterator
func (l *MapSequence) Iterator() Iterable {
	return NewMapIterator(l.data)
}

//Parent returns the sequence as a sequencable
func (l *MapSequence) Parent() Sequencable {
	return Sequencable(l)
}

//Value returns a the final value of a sequence operation
func (l *MapSequence) Value() interface{} {
	return l.Obj()
}

//Value returns a the final value of a sequence operation
func (l *ListSequence) Value() interface{} {
	return l.Obj()
}

//Get retrieves the value
func (l *MapSequence) Get(d interface{}) interface{} {
	l.lock.RLock()
	f := l.data[d]
	l.lock.RUnlock()
	return f
}

//Clone copies internal structure data
func (l *MapSequence) Clone() MapSequencable {
	// l.data = make([]interface{}, 0)
	nd := make(map[interface{}]interface{})

	for k, v := range l.data {
		nd[k] = v
	}

	return NewMapSequence(nd, l.buffer)
}

//Clear wipes internal structure data
func (l *MapSequence) Clear() MapSequencable {
	l.data = make(map[interface{}]interface{})
	return l
}

//Length returns length of data
func (l *MapSequence) Length() int {
	l.lock.RLock()
	sz := len(l.data)
	l.lock.RUnlock()
	return sz
}

//Obj returns the sequence data in the format of its input
func (l *MapSequence) Obj() map[interface{}]interface{} {
	l.lock.RLock()
	m := l.data
	l.lock.RUnlock()
	return m
}

//Add for the ListSequence adds all supplied arguments at once to the list
func (l *MapSequence) Add(f ...interface{}) MapSequencable {
	// l.writer.Stack(func() {
	l.lock.Lock()
	key := f[0]
	val := f[1]
	l.data[key] = val
	l.lock.Unlock()
	// })
	// l.writer.Flush()
	return l
}

//Delete for the ListSequence adds all supplied arguments at once to the list
func (l *MapSequence) Delete(f ...interface{}) MapSequencable {
	l.lock.Lock()
	for _, v := range f {
		_, ok := l.data[v]

		if !ok {
			return l
		}

		delete(l.data, v)
	}
	l.lock.Unlock()

	return l
}

//Values returns the values of this sequence as a sequencable
func (l *MapSequence) Values() ListSequencable {
	kl := NewListSequence(nil, 0)
	it := l.Iterator()

	for it.Next() == nil {
		kl.Add(it.Value())
	}

	return kl
}

//Keys returns the root sequence as a sequencable
func (l *MapSequence) Keys() ListSequencable {
	kl := NewListSequence(nil, 0)
	it := l.Iterator()

	for it.Next() == nil {
		kl.Add(it.Key())
	}

	return kl
}

//ListSequence represents a sequence for arrays,splice type structures
type ListSequence struct {
	*Sequence
	data   []interface{}
	buffer int
}

//Mutate allows mutation on sequence data
func (l *ListSequence) Mutate(fn MutFunc) {
	l.lock.Lock()
	res, ok := fn(l.data).([]interface{})

	if !ok {
		return
	}

	l.data = res
	l.lock.Unlock()
}

//Obj returns the sequence data in the format of its input
func (l *ListSequence) Obj() []interface{} {
	l.lock.RLock()
	d := l.data
	l.lock.RUnlock()
	return d
}

//Iterator returns the sequence data iterator
func (l *ListSequence) Iterator() Iterable {
	return NewListIterator(l.data)
}

//Parent returns the sequence as a sequencable
func (l *ListSequence) Parent() Sequencable {
	return Sequencable(l)
}

//Values returns the value of these sequence as a sequencable
func (l *ListSequence) Values() ListSequencable {
	return l
}

//Keys returns the root sequence as a sequencable
func (l *ListSequence) Keys() ListSequencable {
	kl := NewListSequence(nil, 0)
	keys := l.Iterator()

	for keys.Next() == nil {
		kl.Add(keys.Key())
	}

	return kl
}

//Get retrieves the value
func (l *ListSequence) Get(d interface{}) interface{} {
	dd, ok := d.(int)

	if !ok {
		return nil
	}

	l.lock.RLock()
	val := l.data[dd]
	l.lock.RUnlock()

	return val
}

//Clone copies internal structure data
func (l *ListSequence) Clone() ListSequencable {
	// l.data = make([]interface{}, 0)
	nd := make([]interface{}, l.Length())
	copy(nd, l.data)
	return NewListSequence(nd, l.buffer)
}

//Clear wipes internal structure data
func (l *ListSequence) Clear() ListSequencable {
	l.data = make([]interface{}, 0)
	return l
}

//Length returns length of data
func (l *ListSequence) Length() int {
	l.lock.RLock()
	sz := len(l.data)
	l.lock.RUnlock()
	return sz
}

//Add for the ListSequence adds all supplied arguments at once to the list
func (l *ListSequence) Add(f ...interface{}) ListSequencable {
	l.lock.Lock()
	l.data = append(l.data, f...)
	l.lock.Unlock()
	return l
}

//Delete for the ListSequence adds all supplied arguments at once to the list
func (l *ListSequence) Delete(f ...interface{}) ListSequencable {
	if len(l.data) <= 0 {
		return l
	}

	for _, v := range f {

		i, ok := v.(int)

		if !ok {
			return l
		}

		if i <= 0 && len(l.data) <= 0 {
			return l
		}

		l.lock.Lock()
		copy(l.data[i:], l.data[i+1:])
		l.data[len(l.data)-1] = nil
		l.data = l.data[:len(l.data)-1]
		l.lock.Unlock()

	}

	return l
}

//MapIterator provides an iterator for the map structure
type MapIterator struct {
	Iterable
	data    map[interface{}]interface{}
	updater func(*MapIterator)
}

//GrabKeys returns a list of the given map keys
func GrabKeys(b map[interface{}]interface{}) []interface{} {
	keys := make([]interface{}, len(b))
	count := 0

	for k := range b {
		keys[count] = k
		count++
	}

	return keys
}

//NewMapIterator returns a new mapiterator for use
func NewMapIterator(m map[interface{}]interface{}) *MapIterator {
	keys := GrabKeys(m)
	kit := NewListIterator(keys)

	upd := func(f *MapIterator) {
		keys = GrabKeys(f.data)
		f.Iterable = NewListIterator(keys)
	}

	return &MapIterator{Iterable(kit), m, upd}
}

//NewReverseMapIterator returns a new mapiterator for use
func NewReverseMapIterator(m map[interface{}]interface{}) *MapIterator {
	keys := GrabKeys(m)
	kit := NewReverseListIterator(keys)

	upd := func(f *MapIterator) {
		keys = GrabKeys(f.data)
		f.Iterable = NewReverseListIterator(keys)
	}

	return &MapIterator{Iterable(kit), m, upd}
}

//GenerativeIterator is the base iterator for creating custom iterator
//behaviours
type GenerativeIterator struct {
	proc  ProcFunc
	value interface{}
	index interface{}
	can   bool
	count int
}

//NewGenerativeIterator returns a new GenerativeIterator
func NewGenerativeIterator(p ProcFunc) *GenerativeIterator {
	return &GenerativeIterator{
		p,
		nil,
		nil,
		true,
		0,
	}
}

//HasNext calls the next item
func (l *GenerativeIterator) hasNext() bool {
	return l.can
}

//Next moves to the next item
func (l *GenerativeIterator) Next() error {
	if !l.hasNext() {
		return ErrBADINDEX
	}

	v, k, err := l.proc(l)

	if err == ErrBADValue {
		l.value = nil
		l.index = nil
		l.can = false
		return ErrBADValue
	}

	if err == ErrENDINDEX {
		l.can = false
		return err
	}

	l.value = v
	l.index = k
	l.count++
	return err
}

//Reset reverst the iterators index
func (l *GenerativeIterator) Reset() {
	l.value = nil
	l.index = nil
	l.count = 0
}

//Key returns the current index of the iterator
func (l *GenerativeIterator) Key() interface{} {
	return l.index
}

//Value returns the value of the data with the index value
func (l *GenerativeIterator) Value() interface{} {
	return l.value
}

//Length returns the total time this iterator as generated values
func (l *GenerativeIterator) Length() int {
	return l.count
}

//Clone returns a new iterator off that data
func (l *GenerativeIterator) Clone() Iterable {
	return NewGenerativeIterator(l.proc)
}

//BaseIterator handles interation over an iterator
type BaseIterator struct {
	parent Iterable
	value  interface{}
	index  interface{}
	proc   ProcFunc
}

//IdentityIterator takes an Iterable and returns an iterator that simple returns
//the root iterators key and value without change,useful for IteratorSequence
func IdentityIterator(b Iterable) *BaseIterator {
	return NewBaseIterator(b, func(root Iterable) (interface{}, interface{}, error) {
		return root.Value(), root.Key(), nil
	})
}

//NewBaseIterator returns a base iterator based on a function evaluator
func NewBaseIterator(b Iterable, fn ProcFunc) *BaseIterator {
	return &BaseIterator{
		b.Clone(),
		nil,
		nil,
		fn,
	}
}

//Next moves to the next item
func (l *BaseIterator) Next() error {
	err := l.parent.Next()

	if err == ErrBADValue {
		l.value = nil
		l.index = nil
		return ErrBADValue
	}

	if err != nil {
		return err
	}

	v, k, err := l.proc(l.parent)

	if err != nil {
		return err
	}

	l.value = v
	l.index = k
	return nil
}

//Reset reverst the iterators index
func (l *BaseIterator) Reset() {
	l.parent.Reset()
	l.value = nil
	l.index = nil
}

//Key returns the current index of the iterator
func (l *BaseIterator) Key() interface{} {
	return l.index
}

//Value returns the value of the data with the index value
func (l *BaseIterator) Value() interface{} {
	return l.value
}

//Length returns the parent iterators targets length,not its operation length
func (l *BaseIterator) Length() int {
	return l.parent.Length()
}

//Clone returns a new iterator off that data
func (l *BaseIterator) Clone() Iterable {
	return NewBaseIterator(l.parent, l.proc)
}

//ListIterator handles interator over arrays,slices
type ListIterator struct {
	data  []interface{}
	index int
}

//Next moves to the next item
func (m *MapIterator) Next() error {
	err := m.Iterable.Next()
	if m.Iterable.Length() != len(m.data) {
		m.updater(m)
	}
	return err
}

//Value returns the current value of the iterator
func (m *MapIterator) Value() interface{} {
	k := m.Key()
	return m.data[k]
}

//Key returns the current key of the iterator
func (m *MapIterator) Key() interface{} {
	return m.Iterable.Value()
}

//Length returns the iterators targets length,not its operation length
func (m *MapIterator) Length() int {
	return len(m.data)
}

//Clone returns a new iterator off that data
func (m *MapIterator) Clone() Iterable {
	return NewMapIterator(m.data)
}

//ReverseListIterator returns a reverse iterator
type ReverseListIterator struct {
	*ListIterator
}

//NewReverseListIterator returns a new reverse interator
func NewReverseListIterator(b []interface{}) *ReverseListIterator {
	return &ReverseListIterator{NewListIterator(b)}
}

//Key returns the current index of the iterator
func (r *ReverseListIterator) Key() interface{} {
	k, _ := r.ListIterator.Key().(int)

	if k < 0 {
		return nil
	}

	return (len(r.data) - 1) - k
}

//Value returns the value of the data with the index value
func (r *ReverseListIterator) Value() interface{} {
	k, _ := r.Key().(int)

	if k < 0 || k > len(r.data) {
		return nil
	}

	return r.data[k]
}

//Clone returns a new iterator off that data
func (r *ReverseListIterator) Clone() Iterable {
	return NewReverseListIterator(r.data)
}

//NewListIterator returns a new iterator for the []interface{}
func NewListIterator(b []interface{}) *ListIterator {
	return &ListIterator{b, -1}
}

//HasNext calls the next item
func (l *ListIterator) hasNext() bool {
	if len(l.data) > 0 {
		if l.index < 0 || l.index < (len(l.data)-1) {
			return true
		}
	}
	return false
}

//Next moves to the next item
func (l *ListIterator) Next() error {
	if !l.hasNext() {
		return ErrENDINDEX
	}
	l.index++
	return nil
}

//Length returns the iterators targets length,not its operation length
func (l *ListIterator) Length() int {
	return len(l.data)
}

//Clone returns a new iterable off this iterators data
func (l *ListIterator) Clone() Iterable {
	return NewListIterator(l.data)
}

//Reset reverst the iterators index
func (l *ListIterator) Reset() {
	l.index = -1
}

//Key returns the current index of the iterator
func (l *ListIterator) Key() interface{} {
	return l.index
}

//Value returns the value of the data with the index value
func (l *ListIterator) Value() interface{} {
	k, _ := l.Key().(int)

	if k < 0 {
		return nil
	}

	return l.data[k]
}
