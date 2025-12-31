package generate

import (
	"fmt"
	"math/rand"
	"time"
)

// RandomStringCapitalized creates a random string with first letter capitalized
func RandomStringCapitalized(maxLength int) string {
	if maxLength < 3 {
		maxLength = 3
	}

	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	length := rand.Intn(maxLength-2) + 3

	s := make([]rune, length)
	s[0] = rune('A' + rand.Intn(26))
	for i := 1; i < length; i++ {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)
}

// RandomPhoneNumber generates a 10-digit Thai phone number starting with 08 or 09
func RandomPhoneNumber() string {
	prefixes := []string{"08", "09"}
	prefix := prefixes[rand.Intn(len(prefixes))]

	number := make([]rune, 8)
	for i := 0; i < 8; i++ {
		number[i] = rune('0' + rand.Intn(10))
	}

	return prefix + string(number)
}

// RandomEmail generates a random email
func RandomEmail(name string) string {
	domains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com"}
	return fmt.Sprintf("%s%d@%s",
		name,
		rand.Intn(999),
		domains[rand.Intn(len(domains))])
}

// RandomBadmintonScore generates a realistic badminton game score
// Winner gets 21+ points, loser gets less (with deuce handling)
func RandomBadmintonScore() (team1Score, team2Score int, team1Win bool) {
	team1Win = rand.Intn(2) == 0

	// 70% chance of normal game (21-x where x < 20)
	// 30% chance of deuce (22-20, 23-21, etc up to 30-28)
	isDeuce := rand.Float32() < 0.3

	if isDeuce {
		// Deuce scenario: final score between 22-20 and 30-28
		extraPoints := rand.Intn(5) // 0-4 extra points
		winnerScore := 22 + extraPoints
		loserScore := 20 + extraPoints

		if team1Win {
			return winnerScore, loserScore, true
		}
		return loserScore, winnerScore, false
	}

	// Normal game: winner gets 21, loser gets 5-19
	loserScore := rand.Intn(15) + 5 // 5-19

	if team1Win {
		return 21, loserScore, true
	}
	return loserScore, 21, false
}

// RandomBadmintonMatch generates a complete match (best of 3 games)
func RandomBadmintonMatch() (games []struct{ Team1, Team2 int }, team1Wins bool) {
	games = make([]struct{ Team1, Team2 int }, 0, 3)
	team1WinCount := 0
	team2WinCount := 0

	for i := 0; i < 3; i++ {
		t1, t2, t1Win := RandomBadmintonScore()
		games = append(games, struct{ Team1, Team2 int }{t1, t2})

		if t1Win {
			team1WinCount++
		} else {
			team2WinCount++
		}

		// Match over when someone wins 2 games
		if team1WinCount == 2 || team2WinCount == 2 {
			break
		}
	}

	return games, team1WinCount == 2
}

// RandomSkillLevel returns a random skill level based on win rate
func RandomSkillLevel(winRate float64) string {
	if winRate >= 75 {
		return "Expert"
	} else if winRate >= 55 {
		return "Advanced"
	} else if winRate >= 40 {
		return "Intermediate"
	}
	return "Beginner"
}

// RandomSkillPoints calculates skill points based on matches and win rate
func RandomSkillPoints(matches int, winRate float64) int {
	base := matches * 10
	bonus := int(winRate * float64(matches) / 10)
	return base + bonus
}

// RandomDate generates a random date within the last N days
func RandomDate(daysBack int) time.Time {
	secondsBack := rand.Intn(daysBack * 24 * 60 * 60)
	return time.Now().Add(-time.Duration(secondsBack) * time.Second)
}

// ThaiFirstNames common Thai first names
var ThaiFirstNames = []string{
	"Somchai", "Prasert", "Wichai", "Supachai", "Thaworn",
	"Channarong", "Nattapong", "Pichit", "Sompong", "Kittisak",
	"Surasak", "Thanakorn", "Weerachai", "Anon", "Chaiwat",
	"Jirayu", "Pawin", "Teerawat", "Vorapon", "Yuthasak",
}

// ThaiLastNames common Thai last names
var ThaiLastNames = []string{
	"Srisawat", "Chanthong", "Thongsuk", "Rattanakul", "Wongsiri",
	"Pattanasuk", "Siriwan", "Chaiyasit", "Prommas", "Buathong",
	"Saetang", "Jankong", "Suwanrat", "Phonphan", "Yodsaen",
	"Kaewkla", "Thongdee", "Sombun", "Niamhom", "Pankaew",
}

// RandomThaiName generates a random Thai name
func RandomThaiName() (firstName, lastName, nickname string) {
	firstName = ThaiFirstNames[rand.Intn(len(ThaiFirstNames))]
	lastName = ThaiLastNames[rand.Intn(len(ThaiLastNames))]
	// Nickname is often first syllable or shortened
	if len(firstName) > 3 {
		nickname = firstName[:3]
	} else {
		nickname = firstName
	}
	return
}
