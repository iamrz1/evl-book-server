package routes

import (
	"evl-book-server/config"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	FileID      = "profile-picture"
	FilePathKey = "filepath"
)

// UploadPostedImageHandler saves an image to a directory inside the server
// received from the POST request of a multi-part-file
func UploadPostedImageHandler(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	filePath := r.Header.Get(FilePathKey)
	URL := fmt.Sprintf("%s://%s%s/finalize", config.App().Scheme, r.Host, r.URL)

	resp, err := UploadMultipartFile(r, client, URL, FileID, filePath)
	if err != nil {
		_, _ = fmt.Fprint(w, err.Error())
		return
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprint(w, err.Error())
		return
	}

	_, _ = w.Write(bodyText)
}
func UploadMultipartFile(r *http.Request, client *http.Client, uri, key, path string) (*http.Response, error) {
	body, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	mwriter := multipart.NewWriter(writer)
	req.Header.Add("Content-Type", mwriter.FormDataContentType())

	errchan := make(chan error)

	go func() {
		defer close(errchan)
		defer writer.Close()
		defer mwriter.Close()

		w, err := mwriter.CreateFormFile(key, path)
		if err != nil {
			errchan <- err
			return
		}

		in, err := os.Open(path)
		if err != nil {
			errchan <- err
			return
		}
		defer in.Close()

		if written, err := io.Copy(w, in); err != nil {
			errchan <- fmt.Errorf("error copying %s (%d bytes written): %v", path, written, err)
			return
		}

		if err := mwriter.Close(); err != nil {
			errchan <- err
			return
		}
	}()

	for k, v := range r.Header {
		if req.Header.Get(k) == "" {
			req.Header[k] = v
		}
	}

	resp, err := client.Do(req)
	merr := <-errchan

	if err != nil || merr != nil {
		return resp, fmt.Errorf("http error: %v, multipart error: %v", err, merr)
	}

	return resp, nil
}
