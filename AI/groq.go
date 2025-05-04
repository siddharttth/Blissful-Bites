package AI

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

var groqApiKey string

// InitializeGroqModel sets up the Groq API key
func InitializeGroqModel(key string) {
	groqApiKey = key
	log.Println("✅ Groq AI client initialized successfully")
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ImageContent struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type GroqRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// GenGroqAI takes user text data and returns a short diet suggestion
func GenGroqAI(userData string) (string, error) {
	if groqApiKey == "" {
		err := fmt.Errorf("groq AI client not initialized")
		log.Println("❌", err)
		return "", err
	}

	prompt := fmt.Sprintf(`You are Mannat, an expert nutritionist specializing in personalized VEGETARIAN diet plans. Analyze the user's data and provide a concise, actionable vegetarian diet plan. Consider:

1. Current Stats:
   - Weight, height, and BMI
   - Activity level
   - Any health conditions/diseases

2. Goals:
   - Weight management goals (loss/gain/maintenance)
   - Specific fitness objectives
   - Dietary preferences (VEGETARIAN ONLY)

3. History:
   - Past meal tracking data
   - Calorie intake patterns

Provide a personalized VEGETARIAN diet plan that:
- Is realistic and sustainable
- Aligns with their activity level
- Accounts for any health conditions
- Supports their specific goals
- Includes specific meal timing if they're doing intermittent fasting
- Contains NO meat, fish, or seafood (eggs and dairy are acceptable)

Keep the response concise but impactful. Format as:
- Daily calorie target
- Key nutritional focus
- Specific vegetarian meal recommendations

User Data:
%s`, userData)

	request := GroqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	return sendGroqRequest(request)
}

// AnalyzeImageWithGroq takes image data and returns structured food calorie info
func AnalyzeImageWithGroq(imgData []byte, imgType string) (map[string]interface{}, error) {
	if groqApiKey == "" {
		err := fmt.Errorf("groq AI client not initialized")
		log.Println("❌", err)
		return nil, err
	}

	// Convert image to base64
	base64Img := base64.StdEncoding.EncodeToString(imgData)
	dataURI := fmt.Sprintf("data:image/%s;base64,%s", imgType, base64Img)

	prompt := `As a precision nutritionist, analyze this food image and provide detailed nutritional information. Focus on:

1. Identify all visible food items
2. Calculate accurate calorie content for each item
3. Consider portion sizes and preparation methods
4. Account for visible ingredients and likely preparation methods

Provide the analysis in this exact JSON format:
{
    "Food Item 1": calories (integer),
    "Food Item 2": calories (integer),
    ...
    "Total calories": sum_of_all_calories (integer)
}

Requirements:
- Use precise calorie values
- Include ALL visible food items
- Consider serving sizes
- Include ONLY the JSON output, no additional text
- Ensure all calorie values are integers
- Always include the "Total calories" field`

	request := GroqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []Message{
			{
				Role: "user",
				Content: []any{
					TextContent{
						Type: "text",
						Text: prompt,
					},
					ImageContent{
						Type: "image",
						ImageURL: struct {
							URL string `json:"url"`
						}{
							URL: dataURI,
						},
					},
				},
			},
		},
	}

	response, err := sendGroqRequest(request)
	if err != nil {
		return nil, err
	}

	// Parse the response as JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("❌ Failed to parse meal data: %v", err)
		return nil, err
	}

	log.Println("✅ Parsed meal data successfully")
	return result, nil
}

// sendGroqRequest sends a request to the Groq API and returns the response content
func sendGroqRequest(request GroqRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("❌ Failed to marshal request: %v", err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+groqApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read response: %v", err)
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ API request failed with status %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var groqResp GroqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		log.Printf("❌ Failed to unmarshal response: %v", err)
		return "", err
	}

	if len(groqResp.Choices) == 0 {
		log.Println("❌ No response from Groq API")
		return "", fmt.Errorf("no response from Groq API")
	}

	log.Println("✅ Response generated successfully")
	return groqResp.Choices[0].Message.Content, nil
}
