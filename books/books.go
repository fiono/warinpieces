package main

import (
  "bufio"
  "errors"
  "fmt"
  "os"
  "regexp"
)

var bookEnd = regexp.MustCompile("^\\*\\*\\* END OF THIS PROJECT GUTENBERG .+ \\*\\*\\*$")
var chapterHeading = regexp.MustCompile("^PART \\w+$")

func chapterFile(bookId string, chapterInd int) (*os.File, error) {
  return os.Create(fmt.Sprintf("data/%s_%d.txt", bookId, chapterInd))
}

func chapterizeBook(id string) error {
	inFile, err := os.Open(fmt.Sprintf("data/%s.txt", id))
  defer inFile.Close()
  if err != nil {
    return err
  }

  scanner := bufio.NewScanner(inFile)
  chapterInd := 0
  var outFile *os.File

  for scanner.Scan() {
    line := scanner.Text()
    if bookEnd.MatchString(line) {
      return nil
    } else if chapterHeading.MatchString(line) {
      if outFile != nil {
        outFile.Close()
      }

      chapterInd++
      outFile, err = chapterFile(id, chapterInd)
      if err != nil {
        return err
      }
    } else if outFile != nil {
      if _, err := outFile.Write([]byte(scanner.Text() + "\n")); err != nil {
        return err
      }
    }
  }

  return errors.New("Did not find end of book")
}

func main() {
  err := chapterizeBook("1399")
  if err != nil {
    fmt.Println(err)
  }
}
