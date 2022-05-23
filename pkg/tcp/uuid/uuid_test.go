package uuid

import (
	"fmt"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	c := NewUUID32Generator()
	fmt.Println(c.Next())
	fmt.Println(c.Next())
}
