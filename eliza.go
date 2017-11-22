package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Replacer is a struct with two elements: a compiled regular expression,
// as per the regexp package, and an array of strings containing possible
// replacements for a string matching the regular expression.
type Replacer struct {
	original     *regexp.Regexp
	replacements []string
}

// ReadReplacersFromFile reads an array of Replacers from a text file.
// It takes a single argument: a string which is the path to the data file.
// The file should be a series of sections with the following format:
//   All lines that begin with a hash symbol are ignored.
//   Each section should begin with a regular expression on a single line.
//   Each subsequent line, until a blank line, should contain a possible
//   replacement for a string matching the regular expression.
//   Each section should end with at least one blank line.
// The idea is to create an array that can be traversed, looking for the first
// regular expression to match some input string. Once a match is found, a
// random replacement string is returned.
func ReadReplacersFromFile(path string) []Replacer {
	// Open the file, logging a fatal error if it fails, close on return.
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create an empty array of Replacers.
	var replacers []Replacer

	// Read the file line by line.
	for scanner, readoriginal := bufio.NewScanner(file), false; scanner.Scan(); {
		// Read the next line and decide what to do.
		switch line := scanner.Text(); {
		// If the line starts with a # character then skip it.
		case strings.HasPrefix(line, "#"):
			// Do nothing
		// If we see a blank line, then make sure we indicate a new section.
		case len(line) == 0:
			readoriginal = false
		// If we haven't read the original, then append an element to the
		// replacers array, compiling the regular expression. The replacements
		// array is left blank for now.
		case readoriginal == false:
			replacers = append(replacers, Replacer{original: regexp.MustCompile(line)})
			readoriginal = true
		// Otherwise read a replacement and add it to the last replacer.
		default:
			replacers[len(replacers)-1].replacements = append(replacers[len(replacers)-1].replacements, line)
		}
	}
	// Return the replacers array.
	return replacers
}

// Eliza is a data structure representing a chatbot.
// The fields responses and substitutions are arrays of Replacers.
// Eliza will attempt matches from start to end of each array.
type Eliza struct {
	responses     []Replacer
	substitutions []Replacer
}

// ElizaFromFiles reads in text files containing responses and substitutions
// data and returns an instance of Eliza with these loaded in.
func ElizaFromFiles(responsePath string, substitutionPath string) Eliza {
	eliza := Eliza{}

	eliza.responses = ReadReplacersFromFile(responsePath)
	eliza.substitutions = ReadReplacersFromFile(substitutionPath)

	return eliza
}

// RespondTo takes a string as input and returns a string. The returned string
// contains the chatbot's response to the input.
func (me *Eliza) RespondTo(input string) string {
	// Look for a possible response.
	for _, response := range me.responses {
		// Check if the user input matches the original, capturing any groups.
		if matches := response.original.FindStringSubmatch(input); matches != nil {
			// Select a random response.
			output := response.replacements[rand.Intn(len(response.replacements))]
			// We'll tokenise the captured groups using the following regular expression.
			boundaries := regexp.MustCompile(`[\s,.?!]+`)
			// Fill the response with each captured group from the input.
			// This is a bit complex, because we have to reflect the pronouns.
			for m, match := range matches[1:] {
				// First split the captured group into tokens.
				tokens := boundaries.Split(match, -1)
				// Loop through the tokens.
				for t, token := range tokens {
					// Loop through the potential substitutions.
					for _, substitution := range me.substitutions {
						// Check if the original of the current substitution matches the token.
						if substitution.original.MatchString(token) {
							// If it matches, replace the token with one of the replacements (at random).
							// Then break.
							tokens[t] = substitution.replacements[rand.Intn(len(substitution.replacements))]
							break
						}
					}
				}
				// Replace $1 with the first match, $2 with the second, etc.
				// Note that element 0 of matches is the original match, not a captured group.
				output = strings.Replace(output, "$"+strconv.Itoa(m+1), strings.Join(tokens, " "), -1)
			}
			// Send the filled answer back.
			return output
		}
	}
	// If there are no matches, then return this generic response.
	return "I don't know what to say."
}

// Program entry point.
func main() {
	// Seed the rand package with the current time.
	rand.Seed(time.Now().UnixNano())

	// Create a new instance of Eliza.
	eliza := ElizaFromFiles("data/responses.txt", "data/substitutions.txt")

	// Print a greeting to the user.
	fmt.Println("Eliza: Hello, I'm Eliza. How are you feeling today?")
	// Read from the user.
	scanner := bufio.NewScanner(os.Stdin)
	for fmt.Print("You: "); scanner.Scan(); fmt.Print("You: ") {
		// Print Eliza's response.
		fmt.Println("Eliza:", eliza.RespondTo(scanner.Text()))
		// If the user typed "quit" then exit. Eliza has a chance to respond first.
		if quit, _ := regexp.MatchString("(?i)^quit$", scanner.Text()); quit {
			break
		}
	}
}
