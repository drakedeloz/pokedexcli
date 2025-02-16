package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/drakedeloz/pokedexcli/internal/pokecache"
	"github.com/fatih/color"
)

type cliCommand struct {
	Name        string
	Description string
	Callback    func(*Config) error
}

type Config struct {
	Next     *string
	Previous *string
}

type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type LocationAreaResp struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []LocationArea `json:"results"`
}

var cache = pokecache.NewCache(15 * time.Minute)

func main() {
	cfg := &Config{
		Next:     nil,
		Previous: nil,
	}
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	for {
		color.RGB(140, 160, 250).Print("Pokedex > ")
		scanner.Scan()
		userInput := cleanInput(scanner.Text())
		c, ok := commands[userInput[0]]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		c.Callback(cfg)
	}
}

func cleanInput(text string) []string {
	parsedText := strings.TrimSpace(text)
	parsedText = strings.ToLower(parsedText)
	splitText := strings.Fields(parsedText)

	return splitText
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"map": {
			Name:        "map",
			Description: "Get a list of the next 20 location areas in the Pokemon world",
			Callback:    commandMap,
		},
		"mapb": {
			Name:        "mapb",
			Description: "Get a list of the previous 20 location areas in the Pokemon world",
			Callback:    commandMapb,
		},
		"help": {
			Name:        "help",
			Description: "Show all commands",
			Callback:    commandHelp,
		},
		"exit": {
			Name:        "exit",
			Description: "Exit the Pokedex",
			Callback:    commandExit,
		},
	}
}

// COMMAND FUNCTIONS //

func commandExit(cfg *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config) error {
	commands := getCommands()
	fmt.Println("Welcome to the Pokedex!")
	fmt.Print("Usage:\n\n")
	for _, c := range commands {
		fmt.Printf("%s: %s\n", c.Name, c.Description)
	}
	fmt.Print("\n")
	return nil
}

func commandMap(cfg *Config) error {
	var url string
	if cfg.Next == nil {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = *cfg.Next
	}
	err := mapHelper(url, cfg)
	if err != nil {
		return err
	}
	return nil
}

func commandMapb(cfg *Config) error {
	var url string
	if cfg.Previous == nil {
		fmt.Println("You're on the first page.")
		return nil
	} else {
		url = *cfg.Previous
	}
	err := mapHelper(url, cfg)
	if err != nil {
		return err
	}
	return nil
}

////

func mapHelper(url string, cfg *Config) error {
	var resp LocationAreaResp
	var body []byte
	var err error

	if data, found := cache.Get(url); found {
		body = data
		color.RGB(140, 160, 250).Print("Displaying Cached Data\n")
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d", res.StatusCode)
		}
		defer res.Body.Close()
		body, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		cache.Add(url, body)
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	cfg.Next = resp.Next
	cfg.Previous = resp.Previous

	for _, result := range resp.Results {
		fmt.Println(result.Name)
	}

	return nil
}
