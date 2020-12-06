package main

import (
	"bytes"
)

type SizedBufferPool struct {
	c chan *bytes.Buffer
	a int
}

//make a buffer poll which has buffered channel of bytes.Buffer and
func NewSizedBufferPool(size int, alloc int) *SizedBufferPool {
	return &SizedBufferPool{
		c: make(chan *bytes.Buffer, size),
		a: alloc,
	}
}

//get a buffer from the buffer pool or make a new one if a buffer doesn't exist.
//We also give them a preallocate capacity to help make fewer makeSlice calls
func (bp *SizedBufferPool) Get() *bytes.Buffer {
	select {
	case b := <-bp.c:
		return b
	default:
		b := bytes.NewBuffer(make([]byte, 0, bp.a))
		return b
	}
}

//Return the used buffer to the buffer pool
//If the capacity of the buffer is bigger than what we allocated, we recreate it with our regular parms
func (bp *SizedBufferPool) Put(b *bytes.Buffer)  {
	b.Reset()


	if cap(b.Bytes()) > bp.a {
		b = bytes.NewBuffer(make([]byte, 0, bp.a))
	}
	//here we pass the buffer into the pool otherwise(when the pool (chan size) if full) we just discard it
	select {
	case bp.c <- b:
	default:
	}

}
