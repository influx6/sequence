#Sequence
 This is a library providing sequence like structures for go that practically allow any type of operations to be performed without creating intermediary data and allows a vast use cases

##Example

 data = []interface{}{1, 32, 56, 7}
 li := NewListIterator(data)

 for li.HasNext() {
        err := li.Next()

        if err != nil {
	        t.Fatal("Error occcured with reverse list", err)
		break		
	}

	ind, _ := li.Key().(int)
	if li.Value() != data[ind] {
		t.Fatal("Index and value incorrect with list", li.Key(), li.Value(), data)
		break
	}
 }


 //ListSequence structures for list items
 ls := NewListSequence(nil, 0)

 ls.Add(1, 2, 4, 5)

 ls.Delete(2)

 cl := ls.Clone()

 //MapSequence structures for map items
 ms := NewMapSequence(nil, 0) 
 ms.Add(1, 'a')
 ms.Add(2, 'b')
 ms.Add(3, 'c')
 
 ms.Delete(2)
 
 ms.Get(3) //=> 'c'
 
 ##Structures
 All sequence structures in truth work using iterator structures which provide the standard next(),value() and key() function methods to allow retrieval of the current state and these lends itself to be very powerful that apart from the focus structures like ListSequence and MapSequence provide an extendable and powerful approach without the need of intermediate generation of result, this means anything can be turned into a single if it provides an iterator that meets the #Sequence.Iterable interface
 
 
