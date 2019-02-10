package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

var common = []string{
	"123456", "12345", "123456789", "password", "iloveyou",
	"princess", "1234567", "rockyou", "12345678", "abc123",
	"nicole", "daniel", "babygirl", "monkey", "lovely",
	"jessica", "654321", "michael", "ashley", "qwerty",
	"111111", "iloveu", "000000", "michelle", "tigger",
}

type pair struct {
	salt, hash string
}

func (p *pair) fromString(s string) {
	p.salt = s[7:17]
	p.hash = s[18:]
}

func (p *pair) sha1(pw string) string {
	sum := sha1.Sum([]byte(p.salt + pw))
	return hex.EncodeToString(sum[:])
}

var sample = flag.String("sample", "../rockyou-samples.sha1-salt.txt", "sha1-salt sample file")

func main() {
	handle, err := os.Open(*sample)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	r := bufio.NewScanner(handle)
	result := &pair{}
	table := make(map[string]int)
	for r.Scan() {
		result.fromString(r.Text())
		for _, pw := range common {
			hash := result.sha1(pw)

			if hash == result.hash {
				// fmt.Printf("%+v, %s, %s\n", result, hash, pw)
				table[pw]++
				break
			}
		}
	}

	if err := r.Err(); err != nil {
		panic(err)
	}

	handle, err = os.Create("salt-cracked.txt")
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	w := bufio.NewWriter(handle)
	for pw, count := range table {
		w.WriteString(fmt.Sprintf("%d,%s\n", count, pw))
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
}
