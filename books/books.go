package books

import (
  "bufio"
  "errors"
  "fmt"
  "os"
  "regexp"

  "config"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

type BookMeta struct {
  BookId string
  Title string
  Author string
  Chapters int
  Delimiter string
}

// this is gnarly and i should be using the metadata
var bookEnd = regexp.MustCompile("^\\*\\*\\* END OF THIS PROJECT GUTENBERG .+ \\*\\*\\*$")
var authorPatt = regexp.MustCompile("^Author: (.*)$")
var titlePatt = regexp.MustCompile("^Title: (.*)$")

func getBucket(ctx context.Context) (bkt *storage.BucketHandle, err error) {
  cfg := config.LoadConfig()

  client, err := storage.NewClient(ctx)
  if err != nil {
    return nil, err
  }

  return client.Bucket(cfg.Storage.BucketName), nil
}

func ChapterizeBook(bookId string, delimiter string, ctx context.Context) (meta BookMeta, err error) {
  chapterPatt := regexp.MustCompile(fmt.Sprintf("^%s \\w+$", delimiter))

  var author string
  var title string

  path := fmt.Sprintf("test_data/%s.txt", bookId)
	inFile, err := os.Open(path)
  defer inFile.Close()
  if err != nil {
    return meta, err
  }

  scanner := bufio.NewScanner(inFile)
  chapterInd := 0

  bkt, err := getBucket(ctx)
  if err != nil {
    return meta, err
  }
  var objWriter *storage.Writer

  for scanner.Scan() {
    line := scanner.Text()
    if bookEnd.MatchString(line) {
      return BookMeta{bookId, title, author, chapterInd, delimiter}, nil
    } else if authorPatt.MatchString(line) {
      author = authorPatt.FindStringSubmatch(line)[1]
    } else if titlePatt.MatchString(line) {
      title = titlePatt.FindStringSubmatch(line)[1]
    } else if chapterPatt.MatchString(line) {
      if objWriter != nil {
        if err := objWriter.Close(); err != nil {
          return meta, err
        }
      }
      chapterInd++
      objWriter = bkt.Object(fmt.Sprintf("books/%s/%d.txt", bookId, chapterInd)).NewWriter(ctx)
    } else if objWriter != nil {
      if _, err := objWriter.Write([]byte(scanner.Text() + "\n")); err != nil {
        return meta, err
      }
    }
  }

  err = objWriter.Close()
  if err != nil {
    return meta, err
  }
  return meta, errors.New("Did not find end of book")
}
