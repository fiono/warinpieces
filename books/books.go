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

var bookEnd = regexp.MustCompile("^\\*\\*\\* END OF THIS PROJECT GUTENBERG .+ \\*\\*\\*$")

func getBucket(ctx context.Context) (bkt *storage.BucketHandle, err error) {
  cfg := config.LoadConfig()

  client, err := storage.NewClient(ctx)
  if err != nil {
    return nil, err
  }

  return client.Bucket(cfg.Storage.BucketName), nil
}

func ChapterizeBook(bookId string, delimiter string, ctx context.Context) error {
  chapterHeading := regexp.MustCompile(fmt.Sprintf("^%s \\w+$", delimiter))

  path := fmt.Sprintf("test_data/%s.txt", bookId)
	inFile, err := os.Open(path)
  defer inFile.Close()
  if err != nil {
    return err
  }

  scanner := bufio.NewScanner(inFile)
  chapterInd := 0

  bkt, err := getBucket(ctx)
  if err != nil {
    return err
  }
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

  if err := objWriter.Close(); err != nil {
    return err
  }
  return errors.New("Did not find end of book")
}
