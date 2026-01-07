package ui

import (
	"fmt"
	"time"
)

// Spinner represents an animated spinner
type Spinner struct {
	message string
	stop    chan bool
	done    chan bool
	frames  []string
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		stop:    make(chan bool),
		done:    make(chan bool),
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Start starts the spinner animation
func (s *Spinner) Start() {
	go func() {
		i := 0
		for {
			select {
			case <-s.stop:
				// Clear the line
				fmt.Print("\r\033[K")
				s.done <- true
				return
			default:
				frame := s.frames[i%len(s.frames)]
				fmt.Printf("\r%s %s", frame, s.message)
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.stop <- true
	<-s.done
}

// StopWithMessage stops the spinner and shows a completion message
func (s *Spinner) StopWithMessage(message string) {
	s.Stop()
	fmt.Println(message)
}

// ShowDots shows an animated dots loader for a fixed duration
func ShowDots(message string, duration time.Duration) {
	done := make(chan bool)
	go func() {
		time.Sleep(duration)
		done <- true
	}()

	dotCount := 0
	for {
		select {
		case <-done:
			// Clear the line and show completion
			fmt.Printf("\r%s%s\n", message, "...")
			return
		default:
			dots := ""
			for i := 0; i < (dotCount%3)+1; i++ {
				dots += "."
			}
			fmt.Printf("\r%s%s", message, dots)
			time.Sleep(300 * time.Millisecond)
			dotCount++
		}
	}
}

// PrintWithDots executes a function while showing animated dots
func PrintWithDots(message string, fn func() error) error {
	done := make(chan bool)
	var fnErr error

	go func() {
		fnErr = fn()
		done <- true
	}()

	dotCount := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r%s%s\n", message, "...")
			return fnErr
		default:
			dots := ""
			for i := 0; i < (dotCount%3)+1; i++ {
				dots += "."
			}
			fmt.Printf("\r%s%s", message, dots)
			time.Sleep(300 * time.Millisecond)
			dotCount++
		}
	}
}
