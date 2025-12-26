package main

import (
	"backend/database/generate"
	"fmt"
)

func main() {

	var size int
	// get size from user input
	fmt.Print("Please enter number of people to generate: ")
	fmt.Scanln(&size)
	for size <= 0 {
		fmt.Print("Invalid number. Please enter a positive integer: ")
		fmt.Scanln(&size)
	}

	// get size of first name, last name, nickname from user input
	var sizeFirstName, sizeLastName, sizeNickname int
	fmt.Print("Please enter size of first name: ")
	fmt.Scanln(&sizeFirstName)
	fmt.Print("Please enter size of last name: ")
	fmt.Scanln(&sizeLastName)
	fmt.Print("Please enter size of nickname: ")
	fmt.Scanln(&sizeNickname)

	// generate mock data
	fmt.Println("\nGenerated Mock Data:")
	for i := 0; i < size; i++ {
		firstName := generate.RandomStringCapitalized(sizeFirstName)
		lastName := generate.RandomStringCapitalized(sizeLastName)
		nickname := generate.RandomStringCapitalized(sizeNickname)
		phone := generate.RandomPhoneNumber()

		// fmt.Printf("Person %d: %s %s, Nickname: %s, Phone: %s\n",
		//	i+1, firstName, lastName, nickname, phone)
		fmt.Printf("Person %d: %s %s, %s, %s\n", i+1, firstName, lastName, nickname, phone)
	}
}
