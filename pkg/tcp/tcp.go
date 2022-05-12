package tcp

//import (
//	"fmt"
//	"testing"
//)
//
//func TestServer_Start(t *testing.T) {
//	s := NewServer()
//	s.SetOnConnected(func(inSite <-chan *Packet) <-chan *Packet {
//		outSite := make(chan *Packet)
//		go func() {
//			for m := range inSite {
//				outSite <- m
//			}
//		}()
//		return outSite
//	})
//	fmt.Println("end:", s.Start(":8080"))
//}
