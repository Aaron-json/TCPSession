package ringbuffer_test

import (
	"testing"

	"github.com/Aaron-json/TCPSession/internal/pkg/ringbuffer"
)
func checkOrder(buf *ringbuffer.RingBuffer[int], t testing.TB){
        // Test the ring buffer implementation
        nElements := 128
        for i := range nElements {
                err := buf.Write(i)
                if err != nil {
                        t.Fatal("TestOrder: Write error", err)
                }
        }
        
        for i := range nElements {
                res, err:= buf.Read()
                if err != nil {
                        t.Fatal("TestOrder: Read error", err)
                }
                if res != i {
                        t.Fatal("TestOrder: Incorrent sequence")
                }
        }
}

func TestOverflowAndOrder(t *testing.T){
        buf := ringbuffer.NewRingBuffer[int]()
        for range 1_000_000{
                checkOrder(buf, t)
        }
}
