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

const maxChapterLen = 100000

type BookMeta struct {
  BookId string
  Title string
  Author string
  Chapters int
  Delimiter string
  ScheduleType int
}

type SubscriptionMeta struct {
  SubscriptionId string
  BookId string
  Email string
  ChaptersSent int
  Active bool
  Validated bool
}

// this is gnarly and i should be using the metadata
var bookEnd = regexp.MustCompile("^\\*\\*\\* ?END .+\\*\\*\\*$")
var authorPatt = regexp.MustCompile("^Author: (.*)$")
var titlePatt = regexp.MustCompile("^Title: (.*)$")

func getPath(bookId string, chapterInd int) string {
  return fmt.Sprintf("books/%s/%d.txt", bookId, chapterInd)
}

func getBucket(ctx context.Context) (bkt *storage.BucketHandle, err error) {
  cfg := config.LoadConfig()

  client, err := storage.NewClient(ctx)
  if err != nil {
    return
  }

  return client.Bucket(cfg.Storage.BucketName), nil
}

func GetChapter(bookId string, chapter int, ctx context.Context) (body string, err error) {
  bkt, err := getBucket(ctx)
  if err != nil {
    return
  }

  obj := bkt.Object(getPath(bookId, chapter))
  r, err := obj.NewReader(ctx)
  if err != nil {
    return
  }
  defer r.Close()

  buf := make([]byte, maxChapterLen)
  n, err := bufio.NewReader(r).Read(buf)
  if err != nil {
    return
  }

  return string(buf[:n]), nil
}

func ChapterizeBook(bookId string, delimiter string, ctx context.Context) (meta BookMeta, err error) {
  chapterPatt := regexp.MustCompile(fmt.Sprintf("^%s .+$", delimiter))

  var author, title string
  var chapterInd = 0

  path := fmt.Sprintf("test_data/%s.txt", bookId)
	inFile, err := os.Open(path)
  if err != nil {
    return
  }
  defer inFile.Close()

  scanner := bufio.NewScanner(inFile)

  bkt, err := getBucket(ctx)
  if err != nil {
    return
  }

  var objWriter *storage.Writer

  for scanner.Scan() {
    line := scanner.Text()
    if bookEnd.MatchString(line) {
      return BookMeta{bookId, title, author, chapterInd, delimiter, 0}, nil // BIGF
    } else if authorPatt.MatchString(line) && author == "" {
      author = authorPatt.FindStringSubmatch(line)[1]
    } else if titlePatt.MatchString(line) && title == "" {
      title = titlePatt.FindStringSubmatch(line)[1]
    } else if chapterPatt.MatchString(line) {
      chapterInd++

      if objWriter != nil {
        objWriter.Close()
      }
      objWriter = bkt.Object(getPath(bookId, chapterInd)).NewWriter(ctx)
    } else if objWriter != nil {
      if _, err = objWriter.Write([]byte(scanner.Text() + "\n")); err != nil {
        return
      }
    }
  }

  objWriter.Close()
  return meta, errors.New("Did not find end of book")
}
