package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := flag.String("password", "", "Password to hash")
	flag.Parse()

	if *password == "" {
		fmt.Println("Please provide a password using the -password flag")
		flag.Usage()
		os.Exit(1)
	}

	// Generate a salt and hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hashed password: %s\n", string(hashedPassword))
} 