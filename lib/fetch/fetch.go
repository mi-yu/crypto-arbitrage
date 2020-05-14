package fetch

import (
	"io/ioutil"
	"log"
	"net/http"
)

// Fetch makes an API request to specified endpoint, handling error
func Fetch(endpoint string) ([]byte, error) {
	res, err := http.Get(endpoint)
	if err != nil {
		log.Println("Error in Fetch:")
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Println("Error in Fetch:")
		log.Fatal(err)
	}
	return body, err
}
