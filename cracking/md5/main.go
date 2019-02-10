package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

var sample = flag.String("sample", "../rockyou-samples.md5.txt", "md5 sample file")

func buildString(set []byte, indices []int) string {
	val := make([]byte, len(indices))
	for k, i := range indices {
		val[k] = set[i]
	}

	return string(val)
}

// Borrowed from https://stackoverflow.com/a/22739674/1517394
func nAryProduct(input string, n int) []string {
	if n <= 0 {
		return nil
	}

	// Copy input into initial product set -- a set of
	// one character sets
	prod := make([]string, len(input))
	for i, char := range input {
		prod[i] = string(char)
	}

	for i := 1; i < n; i++ {
		// The bigger product should be the size of the input times the size of
		// the n-1 size product
		next := make([]string, 0, len(input)*len(prod))

		// Add each char to each word and add it to the new set
		for _, word := range prod {
			for _, char := range input {
				next = append(next, word+string(char))
			}
		}

		prod = next
	}

	return prod
}

const passwordSize = 5

type pair struct {
	cleartext string
	hashed    string
}

func consumer(in <-chan string, out chan<- pair) {
	for str := range in {
		result := md5.Sum([]byte(str))
		out <- pair{
			str,
			hex.EncodeToString(result[:]),
		}
	}
}

func main() {
	flag.Parse()
	from := time.Now()

	// [0-9a-z] (length = 5)
	sampleFile, err := os.Open(*sample)
	if err != nil {
		panic(err)
	}
	defer sampleFile.Close()

	// Populate character set
	fmt.Print("Populating character set... ")
	characters := make([]byte, 0, 36)
	{
		// Numbers
		for i := 0; i < 10; i++ {
			b := byte(strconv.Itoa(i)[0])
			characters = append(characters, b)
		}

		// Characters
		for i := 97; i < 97+26; i++ {
			b := byte(i)
			// fmt.Printf("%d->%s\n", b, string(b))
			characters = append(characters, b)
		}
	}
	// fmt.Print("%+v", len(characters)
	fmt.Println("done!")

	// Generate password set
	fmt.Print("Generating password set... ")
	t := time.Now()
	passwordSet := nAryProduct(string(characters), passwordSize)
	fmt.Printf("done in %s!\n", time.Now().Sub(t))

	// Generate rainbow table
	fmt.Print("Generating rainbow table... ")
	table := make(map[string]string)

	// Build channels
	const chanCount = 4000
	result := make(chan pair)
	c := make(chan string)

	fmt.Print("consumers created, ")
	for i := 0; i < chanCount; i++ {
		go consumer(c, result)
	}

	// Fan out
	go func() {
		fmt.Printf("fanning out %d passwords...", len(passwordSet))
		for i, clear := range passwordSet {
			c <- clear

			if i%5000000 == 0 {
				fmt.Printf(" %d", len(passwordSet)-i)
			} else if i%1000000 == 0 {
				fmt.Print(".")
			}
		}
		fmt.Printf(", ")

		// Close all channels
		close(c)
	}()

	// Fan in
	fmt.Printf("waiting for fan in... ")
	remaining := len(passwordSet)
	for pair := range result {
		remaining--
		table[pair.hashed] = pair.cleartext

		if remaining == 0 {
			close(result)
		}

		if remaining%5000000 == 0 {
			fmt.Printf("..(%d) .. ", remaining)
		}

	}
	fmt.Printf(", done!")

	// Read file
	fmt.Print("Scanning sample... ")
	scanner := bufio.NewScanner(sampleFile)
	results := make(map[string]int)
	for scanner.Scan() {
		pw := scanner.Text()

		clear, ok := table[pw]
		if !ok {
			// panic(errors.Errorf("Could not find %s in hashset", pw))
			continue // We only want to crack 5 characters passwords. Others will be in this set
		}

		// fmt.Printf("Written %s\n", clear)
		results[clear]++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println("done!")

	// Write results
	fmt.Print("Writing results... ")
	outFile, err := os.Create("md5-cracked.txt")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for pw, n := range results {
		writer.WriteString(fmt.Sprintf("%d,%s\n", n, pw))
	}
	fmt.Println("done!")

	if err := writer.Flush(); err != nil {
		panic(err)
	}

	fmt.Printf("\nDuration: %s", time.Now().Sub(from))
}
