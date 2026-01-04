package tunnel

import (
	"math/rand"
	"time"
)

// Funny quotes to display after tunnel connection
var funnyQuotes = []string{
	"Your local server is now on the internet!",
	"Tunneling through the matrix... successfully!",
	"Localhost? More like Worldhost now!",
	"Your server just got a public address!",
	"From local to global in seconds!",
	"Breaking through firewalls like a pro!",
	"Your app is now accessible worldwide!",
	"Tunnel established. Welcome to the internet!",
	"Local development, global access!",
	"Your server is now public and proud!",
	"From zero to hero in one tunnel!",
	"Making localhost great again!",
	"Your code is now globally accessible!",
	"Tunneling like a boss!",
	"Local server, global reach!",
}

// GetRandomQuote returns a random funny quote
func GetRandomQuote() string {
	rand.Seed(time.Now().UnixNano())
	return funnyQuotes[rand.Intn(len(funnyQuotes))]
}

