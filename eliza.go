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

type substitution struct {
	original   *regexp.Regexp
	substitute string
}
type response struct {
	question *regexp.Regexp
	answers  []string
}

type Eliza struct {
	responses     []response
	substitutions []substitution
}

func (me *Eliza) readsubstitutions(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(bufio.NewReader(file))
	scanner.Split(bufio.ScanLines)

	for readoriginal := false; scanner.Scan(); {
		s := scanner.Text()

		switch {
		case strings.HasPrefix(s, "#") || len(s) == 0:
			// Do nothing
		case readoriginal == false:
			me.substitutions = append(me.substitutions, substitution{original: regexp.MustCompile(s)})
			readoriginal = true
		default:
			me.substitutions[len(me.substitutions)-1].substitute = s
			readoriginal = false
		}
	}
}

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
