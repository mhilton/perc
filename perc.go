package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
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
		var prev uint64
		ra := NewRunningAverage(5)

		for n := range c {
			sum += uint64(n)
			perc := (sum * 100) / size
			for ; prev < perc; prev++ {
				ra.Sample()
			}

			if perc == 0 {
				fmt.Fprintf(os.Stderr, "\r  0%% estimated time remaining --:--:--")
				continue
			}

			remaining := time.Duration(100-perc)*ra.Average() - ra.Since()
			hours := remaining / time.Hour
			remaining -= hours * time.Hour
			mins := remaining / time.Minute
			remaining -= mins * time.Minute
			secs := remaining / time.Second

			fmt.Fprintf(os.Stderr, "\r%3d%% estimated time remaining %02d:%02d:%02d", perc, hours, mins, secs)
		}

		fmt.Fprintln(os.Stderr, "")
	}
}

// copy copies bytes from in to out, updating c with number of bytes copied
// in each iteration.
func copy(in io.Reader, out io.Writer, c chan int) {
	defer close(c)

	b := make([]byte, 8*1024)

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

// A RunningAverage maintains a running average of times between samples
type RunningAverage struct {
	samples []time.Duration
	count   int
	last    time.Time
}

// NewRunningAverage makes a new RunningAverage with size sample points
func NewRunningAverage(size int) *RunningAverage {
	return &RunningAverage{make([]time.Duration, size), 0, time.Now()}
}

// Sample saves a new sample in the RunningAverage
func (r *RunningAverage) Sample() {
	next := time.Now()
	r.samples[r.count%len(r.samples)] = next.Sub(r.last)
	r.count++
	r.last = next
}

// Average works out the current running average
func (r *RunningAverage) Average() time.Duration {
	if r.count == 0 {
		return time.Duration(0)
	}

	var sum time.Duration
	for _, t := range r.samples {
		sum += t
	}

	if r.count < len(r.samples) {
		return sum / time.Duration(r.count)
	} else {
		return sum / time.Duration(len(r.samples))
	}
}

// Since returns the time since the last sample
func (r *RunningAverage) Since() time.Duration {
	return time.Since(r.last)
}
