package main

import (
  "bufio"
  "errors"
  "fmt"
  "os"
  "regexp"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

var ctx = context.Background()

var bookEnd = regexp.MustCompile("^\\*\\*\\* END OF THIS PROJECT GUTENBERG .+ \\*\\*\\*$")
var chapterHeading = regexp.MustCompile("^PART \\w+$")

func chapterizeBook(bookId string, bkt *storage.BucketHandle) error {
	inFile, err := os.Open(fmt.Sprintf("data/%s.txt", bookId))
  defer inFile.Close()
  if err != nil {
    return err
  }

  scanner := bufio.NewScanner(inFile)
  chapterInd := 0
  var objWriter *storage.Writer

  for scanner.Scan() {
    line := scanner.Text()
    if bookEnd.MatchString(line) {
      return nil
    } else if chapterHeading.MatchString(line) {
      if objWriter != nil {
        if err := objWriter.Close(); err != nil {
          return err
        }
      }

      chapterInd++

      objWriter = bkt.Object(fmt.Sprintf("bigf/data/%s_%d.txt", bookId, chapterInd)).NewWriter(ctx)
    } else if objWriter != nil {
      if _, err := objWriter.Write([]byte(scanner.Text() + "\n")); err != nil {
        return err
      }
    }
  }

  return errors.New("Did not find end of book")
}

func main() {
  client, err := storage.NewClient(ctx)
  if err != nil {
    fmt.Println(err)
  }

  bkt := client.Bucket("gutenbits.appspot.com")

  err = chapterizeBook("1399", bkt)
  if err != nil {
    fmt.Println(err)
  }
}
