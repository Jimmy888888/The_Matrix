package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"golang.org/x/term"
)

  

var (

	frameRatePtr = flag.Int("fps", 6, "每秒幀數")
	
	charsetPtr = flag.String("charset", "アァカ嗨路比醬サタナハAマヤャPLラワガザダバパQvrobmlJNイィキシチニヒミリヰギジヂビピウゥクスツヌフムユュルグズブヅプエェケセテネヘメレヱゲゼデベペオォコソトノホモヨョロヲゴゾドボポヴッンабвгдежзийклмнопрстуフхцчшщъыьэюяΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩ가나다라마바사아자차카타파하", "數字雨的字符集")
	
	densityPtr = flag.Int("density", 2, "列密度 (width/density)")

	CHARSET []rune
	
	rnd *rand.Rand // Thread-safe random number generator
)

  

func init() {

	flag.Parse()
	
	CHARSET = []rune(*charsetPtr)
	
	randSource := rand.NewSource(time.Now().UnixNano())
	
	rnd = rand.New(randSource)

}

  
  

// Column represents a single column of falling characters

type Column struct {

	x int
	
	y int
	
	speed int
	
	length int
	
	chars []rune
	
	iteration int

}

  

// NewColumn initializes a column at a specific x position

func NewColumn(x int, height int) *Column {

	c := &Column{
	
		x: x,
		
		speed: rnd.Intn(2) + 1, // Speed between 1 and 2
		
		length: rnd.Intn(height/2) + height/4, // Length between 1/4 and 3/4 of height

	}
	
	if c.length == 0 { // Ensure length is at least 1
	
		c.length = 1
	
	}
	
	c.reset(0, height)
	
	return c

}

  

// reset places the column back at the top with new random attributes

func (c *Column) reset(width, height int) {

	if width > 0 {
	
		c.x = rnd.Intn(width) + 1
	
	}
	
	c.y = -rnd.Intn(height * 2) // Start off-screen
	
	c.iteration = 0
	
	c.chars = make([]rune, c.length)
	
	for i := range c.chars {
	
		c.chars[i] = rune(CHARSET[rnd.Intn(len(CHARSET))])
	
	}

}

  

// update moves the column down and regenerates characters

func (c *Column) update(width, height int) bool {

// Change multiple characters randomly more often

	if rnd.Intn(10) > 2 {
	
		numChanges := rnd.Intn(3) + 1 // Change 1 to 3 characters
		
		for i := 0; i < numChanges; i++ {
		
			c.chars[rnd.Intn(c.length)] = rune(CHARSET[rnd.Intn(len(CHARSET))])
		
		}
	
	}
	  
	c.y += c.speed
	
	c.iteration++
	
	// If the column has completely passed the bottom, return false to indicate it's off-screen
	
	if c.y-c.length > height {
	
		return false
	
	}
	
	return true

}

  

// draw renders the column on the screen

func (c *Column) draw(buf *bytes.Buffer) {

	// Clear the characters that have fallen off the bottom of the column
	// These are the characters that were at the tail of the column in the previous frame
	// and are no longer part of the current column.
	// The previous tail was at c.y - c.speed - c.length + 1.
	// We need to clear 'c.speed' number of characters starting from this position.
	
	for i := 0; i < c.speed; i++ {
	
		yToClear := c.y - c.speed - c.length + 1 + i
		
		if yToClear >= 0 {
		
			fmt.Fprintf(buf, "\033[%d;%dH ", yToClear+1, c.x)
		
		}
	
	}
	
	// Now draw the current column
	
	for i, char := range c.chars {
	
		yPos := c.y - i
		
		if yPos >= 0 {
		
			// Head of the rain is white
			
			if i == 0 {
			
				fmt.Fprintf(buf, "\033[%d;%dH\033[1;37m%c\033[0m", yPos+1, c.x, char)
			
			} else if i < c.length/3 { // Next part is bright green
			
				fmt.Fprintf(buf, "\033[%d;%dH\033[1;32m%c\033[0m", yPos+1, c.x, char)
			
			} else if i < c.length/2 { // Mid-tail: standard green
			
				fmt.Fprintf(buf, "\033[%d;%dH\033[0;32m%c\033[0m", yPos+1, c.x, char)
			
			} else { // Lower-tail: dim green
			
				fmt.Fprintf(buf, "\033[%d;%dH\033[2;32m%c\033[0m", yPos+1, c.x, char)
			
			}
		
		}
	
	}

}

  

func main() {

// --- Terminal Setup ---

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	
	if err != nil {
	
		panic(err)
	
	}
	
	// Restore terminal state on exit
	
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	
	// Hide cursor
	
	fmt.Print("\033[?25l")
	
	defer fmt.Print("\033[?25h")
	
	// Clear screen
	
	fmt.Print("\033[2J")
	
	defer fmt.Print("\033[2J\033[H")
	
	  
	// Handle Ctrl+C gracefully
	
	sigChan := make(chan os.Signal, 1)
	
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
	
		<-sigChan
		
		term.Restore(int(os.Stdin.Fd()), oldState)
		
		fmt.Print("\033[2J\033[H\033[?25h") // Cleanup screen and show cursor
		
		os.Exit(0)
	
	}()
	
	// --- Animation Setup ---
	
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	
	if err != nil {
	
		panic(err)
	
	}
	
	
	columns := make([]*Column, 0)
	
	for i := 0; i < width; i++ {
	
		// Create a column only if a random check passes, making it sparser
		
		if rnd.Intn(*densityPtr*5) == 0 { // Use densityPtr for sparsity
		
			columns = append(columns, NewColumn(i+1, height))
		
		}
	
	}
	
	  
	
	// Goroutine to listen for 'q' to quit
	
	quitChan := make(chan struct{})
	
	go func() {
	
		buf := make([]byte, 1)
		
		for {
		
			os.Stdin.Read(buf)
			
			if buf[0] == 'q' {
			
				close(quitChan)
			
				return
			
			}
		
		}
	
	}()
	
	  
	
	ticker := time.NewTicker(time.Second / time.Duration(*frameRatePtr)) // Use frameRatePtr
	
	defer ticker.Stop()
	
	  
	
	// Helper function for max
	
	max := func(a, b int) int {
	
		if a > b {
		
			return a
		
		}
		
		return b
	
	}
	
	  
	
	// --- Main Loop ---
	
	for {
	
		select {
		
		case <-quitChan:
		
			return
			
		case <-ticker.C:
			
			// On terminal resize, adjust columns
			
			newWidth, newHeight, err := term.GetSize(int(os.Stdout.Fd()))
			
			if err != nil {
			
				panic(err) // Handle error for GetSize
			
			}
			
			  
			
			if newWidth != width || newHeight != height {
			
				// Adjust existing columns
				
				for _, col := range columns {
				
					if col.x > newWidth {
					
						col.reset(newWidth, newHeight) // Reset if outside new width
					
					}
			
				}
			
				// Clear screen on resize
				
				fmt.Print("\033[2J")
				
				width, height = newWidth, newHeight
			
			}
			
			  
			
			// Move cursor to a safe place before drawing
			
			fmt.Print("\033[H")
			
			  
			
			var buf bytes.Buffer // Buffer for drawing
			
			// Update and filter columns
			
			var activeColumns []*Column
			
			for _, col := range columns {
			
				// Only update and draw if the column is still active
				
				if col.update(width, height) {
				
					activeColumns = append(activeColumns, col)
					
					col.draw(&buf) // Pass buffer to draw
				
				}
			
			}
			
			columns = activeColumns
			
			  
			
			// Add new columns to maintain density
			
			for len(columns) < max(1, width/(*densityPtr)) { // Use densityPtr
			
				columns = append(columns, NewColumn(rnd.Intn(width)+1, height))
			
			}
			
			fmt.Print(buf.String()) // Print buffered output
		
		}
	
	}

}
