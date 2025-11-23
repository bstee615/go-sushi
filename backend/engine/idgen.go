package engine

import (
	"fmt"
	"math/rand"
	"time"
)

// Japanese regions for game ID generation
var japaneseRegions = []string{
	"tokyo",
	"kyoto",
	"osaka",
	"hokkaido",
	"okinawa",
	"nara",
	"hiroshima",
	"fukuoka",
	"nagoya",
	"sapporo",
}

// Japanese flowers for game ID generation
var japaneseFlowers = []string{
	"sakura",
	"ume",
	"tsubaki",
	"ajisai",
	"kiku",
	"fuji",
	"botan",
	"ayame",
	"momiji",
	"hasu",
}

// Famous sushi chefs for player name generation
var sushiChefs = []string{
	"Jiro Ono",
}

// Pop culture and anime characters for player name generation
var popCultureCharacters = []string{
	"Naruto",
	"Totoro",
	"Goku",
	"Pikachu",
	"Luffy",
	"Yoshikage Kira",
	"Jotaro Kujoh",
	"Naruto Uzamaki",
	"Gon Freecs",
}

// Historical figures for player name generation
var historicalFigures = []string{
	"Miyamoto Musashi",
	"Oda Nobunaga",
}

// Initialize random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateGameID generates a memorable game ID in the format region-flower-number
func GenerateGameID() string {
	region := japaneseRegions[rand.Intn(len(japaneseRegions))]
	flower := japaneseFlowers[rand.Intn(len(japaneseFlowers))]
	number := rand.Intn(99-10) + 10 // 10-99

	return fmt.Sprintf("%s-%s-%d", region, flower, number)
}

// GeneratePlayerName generates a random player name from famous sushi chefs,
// pop culture characters, or historical figures
func GeneratePlayerName() string {
	// Combine all name lists
	allNames := make([]string, 0)
	allNames = append(allNames, sushiChefs...)
	allNames = append(allNames, popCultureCharacters...)
	allNames = append(allNames, historicalFigures...)

	// Select a random name
	return allNames[rand.Intn(len(allNames))]
}
