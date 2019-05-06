package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Exam struct {
	questions [](map[string]string)
}

func (exam *Exam) add(questionText, answer string) {
	question := make(map[string]string)
	question["questionText"] = questionText
	question["correctAnswer"] = answer
	question["evaluated"] = "NO"
	question["studentAnswer"] = ""
	question["answeredCorrectly"] = "No"
	exam.questions = append(exam.questions, question)
}

func (exam *Exam) evaluateAnswer(index int, answer string) {
	if exam.questions[index]["evaluated"] == "NO" {
		exam.questions[index]["studentAnswer"] = answer
		exam.questions[index]["evaluated"] = "YES"
		if exam.questions[index]["correctAnswer"] == answer {
			exam.questions[index]["answeredCorrectly"] = "YES"
		}
	} else {
		panic("Answer already evaluated.")
	}
}

func (exam *Exam) showReport() {
	fmt.Println("----------Exam Report------------")
	for idx, question := range exam.questions {
		fmt.Println("Question #", idx+1)
		fmt.Println("Question Text:", question["questionText"])
		fmt.Println("Correct Answer:", question["correctAnswer"])
		fmt.Println("Your Answer:", question["studentAnswer"])
		fmt.Println("Answered Correctly?", question["answeredCorrectly"])
		fmt.Println("Evaluated?", question["evaluated"])
		fmt.Println("------------------------------------------")
	}
	fmt.Println("Total Score:", exam.totalScore(), "%")
}

func (exam *Exam) totalScore() float64 {
	correctCount := 0
	for _, question := range exam.questions {
		if question["answeredCorrectly"] == "YES" {
			correctCount++
		}
	}
	return 100.0 * float64(correctCount) / float64(len(exam.questions))
}

func createExamObject() *Exam {
	exam := new(Exam)

	csvFile, err := os.Open("problems.csv")
	defer csvFile.Close()
	if err != nil {
		panic(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))

	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		}
		if error != nil {
			fmt.Println("ERROR", error)
			panic(err)
		}
		exam.add(strings.TrimSpace(line[0]), strings.TrimSpace(line[1]))
	}

	// shuffle questions
	rand.Shuffle(len(exam.questions), func(i, j int) {
		exam.questions[i], exam.questions[j] = exam.questions[j], exam.questions[i]
	})
	return exam
}

func startExam(exam *Exam, channel chan string, writer io.Writer, read io.Reader) {
	reader := bufio.NewReader(read)
	for index, question := range exam.questions {
		fmt.Fprintln(writer, "Question #", index+1)
		fmt.Fprintln(writer, "Question Text:", question["questionText"])
		fmt.Fprintln(writer, "Your Answer:")

		answer, readError := reader.ReadString('\n')
		if readError != nil {
			panic(readError)
		}
		exam.evaluateAnswer(index, strings.TrimSpace(answer))
	}
	channel <- "DONE"
}

func startTimer(numSeconds int, channel chan string) {
	channel <- "START"
	for i := 0; i < numSeconds; i++ {
		channel <- "INPROGRESS"
		time.Sleep(time.Second)
	}
	channel <- "DONE"
}

func main() {
	exam := createExamObject()
	channel := make(chan string, 1)

	go startTimer(300, channel)
	go startExam(exam, channel, os.Stdout, os.Stdin)

	for {
		msg := <-channel
		if msg == "DONE" {
			exam.showReport()
			break
		}
		time.Sleep(time.Millisecond)
	}
}
