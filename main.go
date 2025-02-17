package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
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
	BaseExperience int     `json:"base_experience"`
	Height         int     `json:"height"`
	Name           string  `json:"name"`
	Stats          []Stats `json:"stats"`
	Types          []Types `json:"types"`
	Weight         int     `json:"weight"`
}
type Stat struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Stats struct {
	BaseStat int  `json:"base_stat"`
	Effort   int  `json:"effort"`
	Stat     Stat `json:"stat"`
}
type Type struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Types struct {
	Slot int  `json:"slot"`
	Type Type `json:"type"`
}
type PokemonEncounters struct {
	Pokemon Pokemon `json:"pokemon"`
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
var cache = pokecache.NewCache(15 * time.Minute)
var pokedex = map[string]Pokemon{}

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
		"catch": {
			Name:        "catch",
			Description: "Attempt to catch a Pokemon",
			Callback:    commandCatch,
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
		"inspect": {
			Name:        "inspect",
			Description: "Show details about a Pokemon you have caught",
			Callback:    commandInspect,
		},
		"pokedex": {
			Name:        "pokedex",
			Description: "Show a list of all the Pokemon you have caught",
			Callback:    commandPokedex,
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
		fmt.Println(err)
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

func commandCatch(cfg *Config, param string) error {
	if param == "" {
		fmt.Println("You must specify a Pokemon to catch")
		return nil
	}
	url := "https://pokeapi.co/api/v2/pokemon/" + param
	var resp Pokemon
	pokemonData, err := pokeget.GetResource(cache, url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = json.Unmarshal(pokemonData, &resp)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s", resp.Name)
	time.Sleep(1 * time.Second)
	fmt.Print(".")
	time.Sleep(1 * time.Second)
	fmt.Print(".")
	time.Sleep(1 * time.Second)
	fmt.Print(".\n")

	maxExp := 340.0
	threshold := 100 - ((float64(resp.BaseExperience) / maxExp) * 90)
	chance := float64(rng.Intn(100))
	caught := chance < threshold
	if caught {
		fmt.Printf("%s was caught!\n", resp.Name)
		pokedex[resp.Name] = resp
	} else {
		fmt.Printf("%s escaped!\n", resp.Name)
	}
	return nil
}

func commandInspect(cfg *Config, param string) error {
	if param == "" {
		fmt.Println("You must specify a Pokemon to inspect")
		return nil
	}
	if pokemon, found := pokedex[param]; found {
		fmt.Printf("Name: %s\n", pokemon.Name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)
		fmt.Println("Stats:")
		for _, stat := range pokemon.Stats {
			fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
		}
		fmt.Println("Types:")
		for _, pokemonType := range pokemon.Types {
			fmt.Printf("  -%s\n", pokemonType.Type.Name)
		}
	} else {
		fmt.Println("You have not caught that Pokemon yet")
	}
	return nil
}

func commandPokedex(cfg *Config, param string) error {
	if len(pokedex) == 0 {
		fmt.Println("You have not caught any Pokemon")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for _, pokemon := range pokedex {
		fmt.Printf(" - %s\n", pokemon.Name)
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
