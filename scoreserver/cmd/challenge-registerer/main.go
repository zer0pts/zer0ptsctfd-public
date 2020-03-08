package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/repository"
	"gopkg.in/yaml.v2"
)

type Uploader interface {
	Upload(name string, data []byte) (string, error)
}

type transfershUploader struct {
	url string
}

func (uploader *transfershUploader) Upload(name string, data []byte) (string, error) {
	u, err := url.Parse(uploader.url)
	if err != nil {
		return "", err
	}
	u.Path = filepath.Join(u.Path, name)

	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		return "", errors.New(string(content))
	}

	return string(content), nil
}

type s3Uploader struct {
	uploader *s3manager.Uploader
	bucket   string
}

func (s *s3Uploader) Upload(name string, data []byte) (string, error) {
	res, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(name),
		Body:   bytes.NewReader(data),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", err
	}
	return res.Location, nil
}

type Challenge struct {
	Name          string
	Description   string   `yaml:"description"`
	Flag          string   `yaml:"flag"`
	Category      string   `yaml:"category"`
	Tags          []string `yaml:"tags"`
	Author        string   `yaml:"author"`
	BaseScore     int      `yaml:"base_score"`
	Difficulty    string   `yaml:"difficulty"`
	IsDynamic     bool     `yaml:"is_dynamic"`
	IsQuestionary bool     `yaml:"is_questionary"`
	Host          *string  `yaml:"host"`
	Port          *string  `yaml:"port"`
}

type Challenges struct {
	Challenges map[string]Challenge `yaml:"challenges"`
}

func run() error {
	dbdsn := os.Getenv("DBDSN")
	if dbdsn == "" {
		return fmt.Errorf("Environmental variable DBDSN is requried")
	}

	flag.Usage = func() {
		fmt.Printf("Usage:\n  %s [OPTIONS] [CHALLENGES]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	dir := flag.String("dir", "", "challenges directory path")
	transfersh := flag.String("transfersh", "", "transfer.sh upload url")
	s3bucket := flag.String("bucket", "", "S3 Bucket Name")
	s3region := flag.String("region", "", "S3 Region Name")

	flag.Parse()
	var uploader Uploader

	if transfersh != nil && *transfersh != "" {
		uploader = &transfershUploader{
			url: *transfersh,
		}
	}
	if s3bucket != nil && *s3bucket != "" && s3region != nil && *s3region != "" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(*s3region),
		}))

		uploader = &s3Uploader{
			uploader: s3manager.NewUploader(sess),
			bucket:   *s3bucket,
		}
	}

	if dir == nil || *dir == "" || uploader == nil {
		flag.Usage()
		return nil
	}

	repo, err := repository.New(dbdsn, nil)
	if err != nil {
		return err
	}

	challenges := filepath.Join(*dir, "challenges.yaml")
	if stat, err := os.Stat(challenges); err != nil || !stat.Mode().IsRegular() {
		return fmt.Errorf("%s not found or not a regular file", challenges)
	}

	chals := Challenges{}
	data, err := ioutil.ReadFile(challenges)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, &chals); err != nil {
		return err
	}

	chalNameMap := make(map[string]struct{})
	for _, chal := range flag.Args() {
		chalNameMap[chal] = struct{}{}
	}

	for name, chal := range chals.Challenges {
		if _, ok := chalNameMap[name]; len(chalNameMap) != 0 && !ok {
			continue
		}
		chal.Name = name
		err := registerChallenge(*dir, chal, repo, uploader)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func compress(dir string) ([]byte, error) {
	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}
		header.Name, err = filepath.Rel(dir, path)
		if err != nil {
			panic(err)
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func registerChallenge(dir string, chal Challenge, repo repository.Repository, uploader Uploader) error {
	// testing description
	t, err := template.New(chal.Name).Parse(chal.Description)
	if err != nil {
		return err
	}
	err = t.Option("missingkey=error").Execute(ioutil.Discard, map[string]interface{}{
		"Port": chal.Port,
		"Host": chal.Host,
	})
	if err != nil {
		return err
	}

	id, err := repo.FindChallengeIDByName(chal.Name)
	if err == nil {
		err = repo.UpdateChallengeByName(
			chal.Name,
			chal.Flag,
			chal.Description,
			chal.Category,
			chal.Difficulty,
			chal.Author,
			chal.BaseScore,
			chal.IsDynamic,
			chal.IsQuestionary,
			chal.Host,
			chal.Port,
		)
		if err != nil {
			return err
		}
		log.Printf("UPDATE %s", chal.Name)
	} else {
		id, err = repo.RegisterChallenge(
			chal.Name,
			chal.Flag,
			chal.Description,
			chal.Category,
			chal.Difficulty,
			chal.Author,
			chal.Tags,
			chal.BaseScore,
			chal.IsDynamic,
			chal.IsQuestionary,
			chal.Host,
			chal.Port,
		)
		if err != nil {
			return err
		}
		log.Printf("ADD %s\n", chal.Name)
	}

	// if distfiles exists then upload
	var buf []byte
	archive := filepath.Join(dir, chal.Name, "distfiles")
	st, err := os.Stat(archive)

	if err == nil && st.IsDir() {
		buf, err = compress(archive)
		if err != nil {
			return err
		}
		filename := chal.Name + "_" + hexdigest(buf) + ".tar.gz"
		attachmentURL, err := uploader.Upload(filename, buf)
		if err != nil {
			return err
		}
		err = repo.AddAttachment(id, attachmentURL)
		if err != nil {
			return err
		}
		log.Printf("UPLOAD %s as %s\n", filename, attachmentURL)
	}

	archive = filepath.Join(dir, chal.Name, "distarchive")
	st, err = os.Stat(archive)
	if err == nil && st.IsDir() {
		filepath.Walk(archive, func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() && strings.HasSuffix(path, ".tar.gz") {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					log.Println(err)
					return nil
				}
				filename := fmt.Sprintf("%s_%s.tar.gz", filepath.Base(info.Name()), hexdigest(data))
				attachmentURL, err := uploader.Upload(filename, data)
				if err != nil {
					log.Println(err)
					return nil
				}
				err = repo.AddAttachment(id, attachmentURL)
				if err != nil {
					log.Println(err)
					return nil
				}
				log.Printf("UPLOAD %s as %s\n", info.Name(), attachmentURL)
				return nil
			}
			return nil
		})
	}

	return nil
}

func hexdigest(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
