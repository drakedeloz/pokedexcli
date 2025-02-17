package pokeget

import (
	"io"
	"log"
	"net/http"

	"github.com/drakedeloz/pokedexcli/internal/pokecache"
	"github.com/fatih/color"
)

func GetResource(cache *pokecache.Cache, url string) ([]byte, error) {
	var body []byte
	if data, found := cache.Get(url); found {
		body = data
		color.RGB(140, 160, 250).Print("Displaying Cached Data\n")
	} else {
		res, err := http.Get(url)
		if err != nil {
			return []byte{}, err
		}
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d", res.StatusCode)
		}
		defer res.Body.Close()
		body, err = io.ReadAll(res.Body)
		if err != nil {
			return []byte{}, err
		}
		cache.Add(url, body)
	}
	return body, nil
}
