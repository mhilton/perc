package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

func main() {
	var size uint64
	var err error

	if len(os.Args) > 1 {
		size, err = strconv.ParseUint(os.Args[1], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid size: %s\n", err)
		}
	}

	c := make(chan int, 1)

	go copy(os.Stdin, os.Stdout, c)

	if size == 0 {
		// If we don't have a size just copy
		for _ = range c {
		}
	} else {
		var sum uint64
		for n := range c {
			sum += uint64(n)
			perc := (sum * 100) / size
			fmt.Fprintf(os.Stderr, "\r%3d%%", perc)
		}
		fmt.Fprintln(os.Stderr, "")
	}
}

// copy copies bytes from in to out, updating c with number of bytes copied 
// in each iteration. 
func copy(in io.Reader, out io.Writer, c chan int) {
	defer close(c)

	b := make([]byte, 8 * 1024)

	for {
		n, err := in.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Fprintf(os.Stderr, "Input error: %s\n", err)
				os.Exit(1)
			}
		}

		_, err = out.Write(b[:n]) 
		if err != nil {
			fmt.Fprintf(os.Stderr, "Output error: %s\n", err)
			os.Exit(1)
		}

		c <- n
	}
}
