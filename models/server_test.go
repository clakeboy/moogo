package models

import (
	"fmt"
	"math"
	"testing"
)

func TestMath(t *testing.T) {
	a := 19.9
	b := 3000.0
	fmt.Println(a * b)
	fmt.Println(math.Ceil(a * b))
}
