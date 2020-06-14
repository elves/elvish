package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

func main() {
	freq := map[int]int{}
	rd := bufio.NewReader(os.Stdin)
	for {
		r, _, err := rd.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		if r > 0x7f {
			freq[int(r)]++
		}
	}
	var keys []int
	for k := range freq {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Printf("%d U+%04d %s\n", freq[k], k, string(rune(k)))
	}
}
