package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func BindByteResponseToStruct(data []byte, responseStruct interface{}) error {
	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, responseStruct); err != nil {
		return errors.New("failed to bind byte response to struct: " + err.Error())
	}

	return nil
}

func BindHttpResponseToStruct(response *http.Response, responseTypeStruct interface{}) error {
	if response == nil {
		return errors.New("response is not valid")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err := BindByteResponseToStruct(body, &responseTypeStruct); err != nil {
		return err
	}

	return nil
}