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
)

// Replacer is a data structure with two elements: a compiled regular expression as per the regexp package
// and an array of strings containing possible replacements for a string mathcing the regular expression.
type Replacer struct {
	original     *regexp.Regexp
	replacements []string
}

// Eliza is a data structure representing a psychoanalyst.
// responses and substitutions are arrays of Replacers.
// The order of the elements in both arrays matters - the responses and substitutions are matched in order.
type Eliza struct {
	responses     []Replacer
	substitutions []Replacer
}

// ReadReplacersFromFile reads an array of Replacers from a text file.
// It takes a single argument, a string which is the path to the data file.
// The file should have the following format:
//   All lines that begin with a hash symbol are comments, and are ignored.
//   Each section of the file should begin with at least one blank line.
//   The next line should be a regular expression.
//   Each subsequent line, until a blank line, should contain a possible
//   replacement for a string matching the regular expression.
func ReadReplacersFromFile(path string) []Replacer {
	// Open the file, logging a fatal error if it fails, close on return.
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create an empty array of Replacers.
	replacers := []Replacer{}

	// Read in the file, adding Replacers to the array.
	for scanner, readoriginal := bufio.NewScanner(file), false; scanner.Scan(); {
		// Decide what to do with the line.
		switch line := scanner.Text(); {
		// If the line is blank or starts with a # character then skip it.
		case strings.HasPrefix(line, "#") || len(line) == 0:
			// Do nothing
			// If we haven't read the original, then append an element to the substitutions array.
		// The regualr expression is compiled, and the substitution is left blank for now.
		case readoriginal == false:
			replacers = append(replacers, Replacer{original: regexp.MustCompile(line)})
			readoriginal = true
		// Otherwise read the substitution and assign it to the last element of the substitutions array.
		default:
			replacers[len(replacers)-1].replacements = append(replacers[len(replacers)-1].replacements, line)
			readoriginal = false
		}
	}

	return replacers
}

// ElizaFromFiles reads in text files containing responses and substitutions data.
func ElizaFromFiles(responsePath string, substitutionPath string) Eliza {
	eliza := Eliza{}

	eliza.responses = ReadReplacersFromFile(responsePath)
	eliza.substitutions = ReadReplacersFromFile(substitutionPath)

	return eliza
}

// RespondTo take a string as input and returns a string containing what Eliza says when given that string as input.
func (me *Eliza) RespondTo(input string) string {
	// Look for a possible response.
	for _, response := range me.responses {
		// Check if the user input matches the original, capturing any groups.
		if matches := response.original.FindStringSubmatch(input); matches != nil {
			// Select a random answer.
			output := response.replacements[rand.Intn(len(response.replacements))]
			// We'll tokenise the captured groups using the following regular expression.
			boundaries := regexp.MustCompile(`\b`)
			// Fill the answer with the captured groups from the matches.
			for m, match := range matches[1:] {
				// Split the captured group into tokens.
				tokens := boundaries.Split(match, -1)
				// Loop through the tokens.
				for t, token := range tokens {
					// If the token matches a substitution, then substitute it and break.
					for _, substitution := range me.substitutions {
						if substitution.original.MatchString(token) {
							tokens[t] = substitution.replacements[rand.Intn(len(substitution.replacements))]
							break
						}
					}
					output = strings.Replace(output, "$"+strconv.Itoa(m+1), strings.Join(tokens, ""), -1)
				}
			}
			// Send the filled answer back.
			return output
		}
	}
	// If there are no matches, then return this generic phrase.
	return "I don't know what to say."
}

// Program entry point.
func main() {
	// Create a new instance of Eliza.
	eliza := ElizaFromFiles("data/responses.txt", "data/substitutions.txt")

	// Print a greeting to the user.
	fmt.Println("Eliza: Hello, I'm Eliza. How are you feeling today?")
	// Read from the user.
	scanner := bufio.NewScanner(os.Stdin)
	for fmt.Print("You: "); scanner.Scan(); fmt.Print("You: ") {
		fmt.Println("Eliza: ", eliza.RespondTo(scanner.Text()))
		if strings.Compare(strings.ToLower(strings.TrimSpace(scanner.Text())), "quit") == 0 {
			break
		}
	}
}
