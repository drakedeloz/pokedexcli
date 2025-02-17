package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/drakedeloz/pokedexcli/internal/pokecache"
	"github.com/drakedeloz/pokedexcli/internal/pokeget"
	"github.com/fatih/color"
)

type cliCommand struct {
	Name        string
	Description string
	Callback    func(*Config, string) error
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

type LocationPokemon struct {
	Name              string              `json:"name"`
	PokemonEncounters []PokemonEncounters `json:"pokemon_encounters"`
}
type Pokemon struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type PokemonEncounters struct {
	Pokemon Pokemon `json:"pokemon"`
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
		param := ""
		if len(userInput) > 1 {
			param = userInput[1]
		}
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		c.Callback(cfg, param)
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
		"explore": {
			Name:        "explore",
			Description: "Explore an area to find all the Pokemon available there",
			Callback:    commandExplore,
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

func commandExit(cfg *Config, param string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config, param string) error {
	commands := getCommands()
	fmt.Println("Welcome to the Pokedex!")
	fmt.Print("Usage:\n\n")
	for _, c := range commands {
		fmt.Printf("%s: %s\n", c.Name, c.Description)
	}
	fmt.Print("\n")
	return nil
}

func commandMap(cfg *Config, param string) error {
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

func commandMapb(cfg *Config, param string) error {
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

func commandExplore(cfg *Config, param string) error {
	if param == "" {
		fmt.Println("You must specify a location to explore")
		return nil
	}
	url := "https://pokeapi.co/api/v2/location-area/" + param
	var resp LocationPokemon
	var body []byte
	var err error

	body, err = pokeget.GetResource(cache, url)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", param)
	fmt.Println("Found Pokemon:")

	for _, item := range resp.PokemonEncounters {
		fmt.Printf(" - %s\n", item.Pokemon.Name)
	}
	return nil
}

////

func mapHelper(url string, cfg *Config) error {
	var resp LocationAreaResp
	var err error

	body, err := pokeget.GetResource(cache, url)
	if err != nil {
		return err
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
