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

// A data structure representing a term that should be replaced in a string.
// original is a regular expression to be matched, and substitute is a string to replace the match with.
// An example use is to replace the word you with the word me.
type substitution struct {
	original   *regexp.Regexp
	substitute string
}

// A data structure representing a user input and a list of responses to it that Eliza can give.
// question is a regular expression representing the user input.
// answers is an array of strings, any of which is a reasonable response to question.
// question can capture groups of characters, and elements of answers can use them.
// $1 is the first match, $2 the second, etc.
type response struct {
	question *regexp.Regexp
	answers  []string
}

// A data structure representing Eliza.
// responses is an array containing elements of type response, as above.
// Likewise, substitutions is an array containing elements of type substitution.
// The order of the elements in both arrays matters - the responses and substitutions are matched in order.
type Eliza struct {
	responses     []response
	substitutions []substitution
}

// Method to read in a text file containing substitutions data.
// It takes a single argument, a string, which is the path to the substitutions data file.
// The file should have the following format:
//   All lines that begin with a hash symbol are comments, and are ignored.
//   Each section of the file should begin with at least one blank line.
//   The next line should be a regular expression for what to substitute.
//   The next line should be the new text for the substitution.
//   After that, there should be at least one blank.
// An example substitutions file is given in data/substitutions.txt.
func (me *Eliza) readsubstitutions(path string) {

	// Open the file, logging a fatal error if it fails.
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	// Set up a line-by-line scanner for the file.
	scanner := bufio.NewScanner(bufio.NewReader(file))
	scanner.Split(bufio.ScanLines)

	// Read through the file line by line.
	// readoriginal is false if we have not yet read the regular expression to match.
	// It is true if we have read the regular expression, and are now looking for the substitution string.
	for readoriginal := false; scanner.Scan(); {
		// Get the text on the current line.
		s := scanner.Text()

		// Decide what to do with the line.
		switch {
			// If the line is blank or starts with a # character then skip it.
			case strings.HasPrefix(s, "#") || len(s) == 0:
				// Do nothing

			// If we haven't read the original, then append an element to the substitutions array.
			// The regualr expression is compiled, and the substitution is left blank for now.
			case readoriginal == false:
				me.substitutions = append(me.substitutions, substitution{original: regexp.MustCompile(s)})
				readoriginal = true
			// Otherwise read the substitution and assign it to the last element of the substitutions array.
			default:
				me.substitutions[len(me.substitutions)-1].substitute = s
				readoriginal = false
		}
	}
}

// Function to read in a text file containing responses data.
// The file should have the following format:
// All lines that begin with a hash symbol are comments, and are ignored.
// This file should have the following format:
//   Each section of the file should begin with at least one blank line.
//   The next line should be a regular expression, matching a user input.
//   Each subsequent line, until a blank line, should contain a response to
//   the usr input. One of these will be chosen at random upon user input.
//   After the responses, there should be at least one blank.
// An example responses file is given in data/responses.txt.
func (me *Eliza) readresponses(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(bufio.NewReader(file))
	scanner.Split(bufio.ScanLines)

	for newsection := true; scanner.Scan(); {
		s := scanner.Text()

		switch {
		case strings.HasPrefix(s, "#"):
			// Do nothing
		case len(s) == 0:
			newsection = true
		case newsection == true:
			me.responses = append(me.responses, response{question: regexp.MustCompile(s)})
			newsection = false
		default:
			me.responses[len(me.responses)-1].answers = append(me.responses[len(me.responses)-1].answers, s)
		}
	}
}

func (me *Eliza) analyse(userinput string) string {
	// Loop through the responses, looking for a match for the user input.
	for _, response := range me.responses {
		if matches := response.question.FindStringSubmatch(userinput); matches != nil {

			// Select a random answer.
			answer := response.answers[rand.Intn(len(response.answers))]

			// Fill the answer with the captured groups from the matches.
			for i, match := range matches[1:] {
				// Reflect the pronouns in the captured group.
				for _, sub := range me.substitutions {
					match = sub.original.ReplaceAllString(match, sub.substitute)
					// Remove any spaces at the start or end.
					match = strings.TrimSpace(match)
				}
				// Replace $1 with the first reflected captured group, $2 with the second, etc.
				answer = strings.Replace(answer, "$"+strconv.Itoa(i+1), match, -1)
			}

			// Clear any ~~ markers from the string. They prevent future matches.
			answer = strings.Replace(answer, "~~", "", -1)

			// Send the filled answer back.
			return answer
		}
	}

	return "I don't know what to say."
}

func main() {
	eliza := Eliza{}

	eliza.readsubstitutions("data/substitutions.txt")
	eliza.readresponses("data/responses.txt")

	fmt.Println("Hello, I'm Eliza. How are you feeling today?")

	for reader := bufio.NewReader(os.Stdin); ; {
		fmt.Print("> ")
		userinput, _ := reader.ReadString('\n')
		userinput = strings.Trim(userinput, "\r\n")

		fmt.Println(eliza.analyse(userinput))

		if strings.Compare(strings.ToLower(strings.TrimSpace(userinput)), "quit") == 0 {
			break
		}
	}
}
