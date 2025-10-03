package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Kost0/L3/internal/repository"
)

func getChatID(username string, token string) (int64, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=@%s", token, username)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			ID int64 `json:"chat_id"`
		} `json:"result"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if !result.OK {
		return 0, fmt.Errorf("failed to get the chat ID")
	}

	return result.Result.ID, nil
}

func sendTG(notify *repository.Notify) error {
	token := os.Getenv("TG_TOKEN")

	chatID, err := getChatID(notify.TGUser, token)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    notify.Text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post("https://api.telegram.org/bot%s/sendMessage", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(string(body))
	}

	return nil
}
