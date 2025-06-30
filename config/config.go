package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	RequiredAPIKey string
	FrontendOrigin string
	ResumeAgentURL string
	Port           string
}

var Config AppConfig

func LoadConfig() {
	if err := checkENV(); err != nil {
		err = godotenv.Load()
		if err != nil {
			log.Println(" (-) file .env not found, using global variable")
		}

		err = checkENV()
		if err != nil {
			log.Panic(err)
		}
	}

	Config.RequiredAPIKey = os.Getenv("API_KEY")
	if Config.RequiredAPIKey == "" {
		log.Println("API_KEY environment variable not set. API key validation will be skipped.")
	}

	Config.FrontendOrigin = os.Getenv("FRONTEND_ORIGIN")
	if Config.FrontendOrigin == "" {
		Config.FrontendOrigin = "http://localhost:5173" // Default value
	}

	Config.ResumeAgentURL = os.Getenv("RESUME_AGENT_URL")
	if Config.ResumeAgentURL == "" {
		Config.ResumeAgentURL = "http://localhost:8000" // Default placeholder
	}

	Config.Port = os.Getenv("PORT")
	if Config.Port == "" {
		Config.Port = "8080" // Default port
	}

	log.Printf("Configured FRONTEND_ORIGIN: %s", Config.FrontendOrigin)
	log.Printf("Configured RESUME_AGENT_URL: %s", Config.ResumeAgentURL)
	log.Printf("Configured PORT: %s", Config.Port)
}

func checkENV() error {
	log.Println("Checking environment.....")

	_, ok := os.LookupEnv("API_KEY")
	if !ok {
		return errors.New("API_KEY does not exist")
	}

	_, ok = os.LookupEnv("FRONTEND_ORIGIN")
	if !ok {
		return errors.New("FRONTEND_ORIGIN does not exist")
	}

	_, ok = os.LookupEnv("RESUME_AGENT_URL")
	if !ok {
		return errors.New("RESUME_AGENT_URL does not exist")
	}

	_, ok = os.LookupEnv("PORT")
	if !ok {
		return errors.New("PORT does not exist")
	}

	log.Println("Environment is complete.")
	return nil
}
