package main
import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)
func main() {
	b := make([]byte, 16)
	rand.Read(b)
	fmt.Println(hex.EncodeToString(b))
}