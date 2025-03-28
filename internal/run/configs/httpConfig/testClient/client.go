package main

import (
	"Demonstration-Service/internal/Application/Domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func main() {
	baseURL := "http://localhost:8080/api/"
	params := url.Values{}
	params.Add("id", "123")

	fullURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Ошибка при отправке запроса: %v\n", err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Ошибка при закрытие тела ответа")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Ошибка: %s (статус: %d)\n", string(body), resp.StatusCode)
		return
	}

	var order Domain.Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		fmt.Printf("Ошибка при декодировании JSON: %v\n", err)
		return
	}

	fmt.Printf("Получен заказ: %+v\n", order)
}
