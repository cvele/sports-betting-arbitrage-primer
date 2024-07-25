package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/bxcodec/faker/v3"
)

// Define the structure for odds
type Odds struct {
	Win  float64 `json:"win"`
	Draw float64 `json:"draw"`
	Lose float64 `json:"lose"`
}

// Define the structure for a game
type Game struct {
	ID      string `json:"id"`
	TeamA   string `json:"team_a"`
	TeamB   string `json:"team_b"`
	Odds    Odds   `json:"odds"`
	EventAt string `json:"event_at"`
}

// Define the structure for a bookmaker
type Bookmaker struct {
	Name  string `json:"name"`
	Games []Game `json:"games"`
}

// Generate random odds
func generateOdds() Odds {
	rand.Seed(time.Now().UnixNano())
	return Odds{
		Win:  roundToTwoDecimal(rand.Float64()*2 + 1),
		Draw: roundToTwoDecimal(rand.Float64()*3 + 2),
		Lose: roundToTwoDecimal(rand.Float64()*4 + 2),
	}
}

// Round a float to two decimal places
func roundToTwoDecimal(val float64) float64 {
	return float64(int(val*100)) / 100
}

// Generate a list of fake games
func generateGames(numGames int) []Game {
	var games []Game
	for i := 0; i < numGames; i++ {
		game := Game{
			ID:      faker.UUIDDigit(),
			TeamA:   faker.Word(),
			TeamB:   faker.Word(),
			Odds:    generateOdds(),
			EventAt: faker.Date(),
		}
		games = append(games, game)
	}
	return games
}

// Generate a list of bookmakers with games using goroutines and channels
func generateBookmakers(numBookmakers, numGamesPerBookmaker int) []Bookmaker {
	var bookmakers []Bookmaker
	var wg sync.WaitGroup
	bookmakerCh := make(chan Bookmaker, numBookmakers)

	for i := 0; i < numBookmakers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bookmaker := Bookmaker{
				Name:  faker.DomainName(),
				Games: generateGames(numGamesPerBookmaker),
			}
			bookmakerCh <- bookmaker
		}()
	}

	go func() {
		wg.Wait()
		close(bookmakerCh)
	}()

	for bookmaker := range bookmakerCh {
		bookmakers = append(bookmakers, bookmaker)
	}

	return bookmakers
}

// Write bookmakers data to a JSON file
func writeBookmakersToFile(bookmakers []Bookmaker, filename string) error {
	data, err := json.MarshalIndent(bookmakers, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// Read bookmakers data from a JSON file
func readBookmakersFromFile(filename string) ([]Bookmaker, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var bookmakers []Bookmaker
	err = json.Unmarshal(data, &bookmakers)
	return bookmakers, err
}

// Calculate the arbitrage percentage for a set of odds
func calculateArbitragePercentage(odds Odds) float64 {
	return (1 / odds.Win) + (1 / odds.Draw) + (1 / odds.Lose)
}

// Calculate the stake allocation for an arbitrage opportunity
func calculateStakes(odds Odds, totalBet float64) (winStake, drawStake, loseStake float64) {
	arbitragePercentage := calculateArbitragePercentage(odds)
	winStake = (totalBet / arbitragePercentage) / odds.Win
	drawStake = (totalBet / arbitragePercentage) / odds.Draw
	loseStake = (totalBet / arbitragePercentage) / odds.Lose
	return winStake, drawStake, loseStake
}

// Find the best odds for each game across different bookmakers
func findBestOdds(bookmakers []Bookmaker) map[string]Odds {
	bestOdds := make(map[string]Odds)
	for _, bookmaker := range bookmakers {
		for _, game := range bookmaker.Games {
			currentBest, exists := bestOdds[game.ID]
			if !exists || game.Odds.Win > currentBest.Win {
				currentBest.Win = game.Odds.Win
			}
			if !exists || game.Odds.Draw > currentBest.Draw {
				currentBest.Draw = game.Odds.Draw
			}
			if !exists || game.Odds.Lose > currentBest.Lose {
				currentBest.Lose = game.Odds.Lose
			}
			bestOdds[game.ID] = currentBest
		}
	}
	return bestOdds
}

// Find arbitrage opportunities among a list of games
func findArbitrageOpportunities(bookmakers []Bookmaker) {
	bestOdds := findBestOdds(bookmakers)
	for gameID, odds := range bestOdds {
		arbitragePercentage := calculateArbitragePercentage(odds)
		if arbitragePercentage < 1 {
			fmt.Printf("Arbitrage opportunity found for game %s\n", gameID)
			fmt.Printf("Odds: Win: %.2f, Draw: %.2f, Lose: %.2f\n", odds.Win, odds.Draw, odds.Lose)
			totalBet := 100.0 // Example total bet amount
			winStake, drawStake, loseStake := calculateStakes(odds, totalBet)
			fmt.Printf("Stakes: Win: %.2f, Draw: %.2f, Lose: %.2f\n", winStake, drawStake, loseStake)
			totalStake := winStake + drawStake + loseStake
			guaranteedProfit := (totalBet / arbitragePercentage) - totalStake
			fmt.Printf("Guaranteed profit: %.2f\n", guaranteedProfit)
			fmt.Println()
		}
	}
}

func main() {
	const filename = "bookmakers.json"
	numBookmakers := 100          // Number of bookmakers
	numGamesPerBookmaker := 10000 // Number of games per bookmaker

	var bookmakers []Bookmaker
	var err error

	if _, err = os.Stat(filename); os.IsNotExist(err) {
		bookmakers = generateBookmakers(numBookmakers, numGamesPerBookmaker)
		if err := writeBookmakersToFile(bookmakers, filename); err != nil {
			fmt.Println("Error writing bookmakers to file:", err)
			return
		}
	} else {
		bookmakers, err = readBookmakersFromFile(filename)
		if err != nil {
			fmt.Println("Error reading bookmakers from file:", err)
			return
		}
	}

	findArbitrageOpportunities(bookmakers)
}
