package AI

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var client *genai.Client

// InitializeModel sets up the Gemini AI model client
func InitializeModel(key string) {
	ctx := context.Background()
	var err error
	client, err = genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		log.Fatalf("❌ Failed to initialize Gemini client: %v", err)
	}
	log.Println("✅ Gemini AI client initialized successfully")
}

// ListModels prints the available models from the generative AI API
func ListModels() error {
	if client == nil {
		return fmt.Errorf("AI client not initialized")
	}

	ctx := context.Background()
	modelIterator := client.ListModels(ctx)
	log.Println("📋 Listing available models:")

	for {
		model, err := modelIterator.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("❌ Error listing models: %v", err)
			return err
		}
		log.Printf("🔹 Model: %s\n", model.Name)
	}
	return nil
}

// GenImageAI takes image data and returns structured food calorie info
func GenImageAI(imgData []byte, imgType string) (map[string]interface{}, error) {
	if client == nil {
		err := fmt.Errorf("AI client not initialized")
		log.Println("❌", err)
		return nil, err
	}

	vmodel := client.GenerativeModel("gemini-pro-vision")
	ctx := context.Background()

	prompt := `
You are an expert in nutritionist where you need to see the food items from the image
and calculate the total calories, also provide the details of every food items with calories intake
is below JSON format. 
NOTE: Don't include anything apart from below format, like number of food items or irrelevant text )
NOTE: Don't use bold words
NOTE: Don't include "'''JSON" in the output
{
	"Food Item 1": no of calories (integer)
	"Food Item 2": no of calories (integer)
	"Total calories": (integer)
}
`
	resp, err := vmodel.GenerateContent(ctx, genai.Text(prompt), genai.ImageData(imgType, imgData))
	if err != nil {
		log.Printf("❌ GenImageAI GenerateContent error: %v", err)
		return nil, err
	}

	log.Println("✅ Image content generated successfully")

	var jsonData []byte
	if jsonData, err = json.Marshal(resp); err != nil {
		log.Printf("❌ Failed to marshal AI response: %v", err)
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Printf("❌ Failed to unmarshal AI response: %v", err)
		return nil, err
	}

	// Extract result string
	value, ok := data["Candidates"].([]interface{})[0].(map[string]interface{})["Content"].(map[string]interface{})["Parts"].([]interface{})[0].(string)
	if !ok {
		log.Println("❌ Failed to extract result from AI response")
		return nil, fmt.Errorf("invalid AI response format")
	}

	var meal map[string]interface{}
	if err := json.Unmarshal([]byte(value), &meal); err != nil {
		log.Printf("❌ Failed to parse meal data: %v", err)
		return nil, err
	}

	log.Println("✅ Parsed meal data successfully")
	return meal, nil
}

// GenAI takes user text data and returns a short diet suggestion
func GenAI(userData string) (string, error) {
	if client == nil {
		err := fmt.Errorf("AI client not initialized")
		log.Println("❌", err)
		return "", err
	}

	model := client.GenerativeModel("models/gemini-1.5-pro-001")
	ctx := context.Background()

	prompt := fmt.Sprintf(`
You are an expert nutritionist, named "Blissful Bites", where you assess users data (along with the calories if they have been tracking) and suggest very very short diet plans (in a nutshell, in few lines~) accordingly.
%s
`, userData)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("❌ GenAI GenerateContent error: %v", err)
		return "", err
	}

	log.Println("✅ User content generated successfully")

	var jsonData []byte
	if jsonData, err = json.Marshal(resp); err != nil {
		log.Printf("❌ Failed to marshal AI response: %v", err)
		return "", err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Printf("❌ Failed to unmarshal AI response: %v", err)
		return "", err
	}

	value, ok := data["Candidates"].([]interface{})[0].(map[string]interface{})["Content"].(map[string]interface{})["Parts"].([]interface{})[0].(string)
	if !ok {
		log.Println("❌ Failed to extract suggestion text from AI response")
		return "", fmt.Errorf("invalid AI response format")
	}

	log.Println("✅ Diet suggestion generated")
	return value, nil
}
