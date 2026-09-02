package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"accounts/internal/helpers/token"
)

func updateOrAppendEnv(filePath, key, value string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(filePath, []byte(fmt.Sprintf("%s=%s\n", key, value)), 0644)
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	found := false
	newLines := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+"=") {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		if len(newLines) > 0 && newLines[len(newLines)-1] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
	}

	return os.WriteFile(filePath, []byte(strings.Join(newLines, "\n")), 0644)
}

func main() {
	saveFlag := flag.Bool("save", true, "Automatically update or append AUTH_TOKEN to .env file")
	envPath := flag.String("env", ".env", "Path to .env file")
	length := flag.Int("length", 32, "Byte length of random entropy (32 bytes = 64 hex characters)")
	flag.Parse()

	generatedToken, err := token.GenerateSecureToken(*length)
	if err != nil {
		log.Fatalf("Error generating auth token: %v", err)
	}

	fmt.Println("==================================================================")
	fmt.Println("             ACCOUNTS gRPC AUTH TOKEN GENERATOR                   ")
	fmt.Println("==================================================================")
	fmt.Printf("Generated Auth Token:\n\n%s\n\n", generatedToken)

	if *saveFlag {
		if err := updateOrAppendEnv(*envPath, "AUTH_TOKEN", generatedToken); err != nil {
			log.Printf("Warning: Failed to write to %s: %v\n", *envPath, err)
		} else {
			fmt.Printf("✔ Successfully saved AUTH_TOKEN to %s\n", *envPath)
		}
	}

	fmt.Println("------------------------------------------------------------------")
	fmt.Println("Usage:")
	fmt.Println("  1. In Postman: Set 'auth_token' collection variable to this value.")
	fmt.Println("  2. In gRPC calls: Pass metadata header 'authorization: Bearer <token>'")
	fmt.Println("==================================================================")
}
