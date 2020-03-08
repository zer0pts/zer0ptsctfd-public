package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Webhook interface {
	Send(text string) error
}

type webhook struct {
	endpoint string
}

func New(endpoint string) Webhook {
	return &webhook{
		endpoint: endpoint,
	}
}

func (w *webhook) Send(text string) error {
	payload, err := json.Marshal(map[string]interface{}{
		"text": text,
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	req, err := http.NewRequest("POST", w.endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer res.Body.Close()

	if !(200 <= res.StatusCode && res.StatusCode < 300) {
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		return errors.New(string(data))
	}

	return nil
}
