package Controllers

import (
	AI "blissfulbites/AI"
	DB "blissfulbites/DB"
	"database/sql"
	"fmt"
	"io/ioutil"
	"math"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// UserDetails struct for database operations
type UserDetails struct {
	Name          string
	Gender        string
	Age           int
	ActivityLevel string
	Goals         string
	Height        float64
	Weight        float64
	TargetWeight  float64
	Diseases      string
	Email         string
	DietPlan      sql.NullString
	Healthscore   int
	Track         []byte
}

func FormHandler(c *gin.Context) {
	fmt.Println("[FormHandler] Starting form submission process...")

	// Parse form data
	err := c.Request.ParseForm()
	if err != nil {
		fmt.Println("[FormHandler] Error parsing form:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	// Get all form values
	formData := make(map[string]interface{})
	for key, values := range c.Request.PostForm {
		if len(values) > 1 {
			formData[key] = values // For multiple values (like checkboxes)
		} else if len(values) == 1 {
			formData[key] = values[0] // For single values
		}
	}

	// Get email from form data
	email := c.PostForm("email")
	if email == "" {
		fmt.Println("[FormHandler] Error: No email provided in form data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	formData["email"] = email

	fmt.Printf("[FormHandler] Processing form data for email: %s\n", email)
	fmt.Printf("[FormHandler] Form data: %+v\n", formData)

	// Insert data into database
	err = DB.InsertUserData(formData)
	if err != nil {
		fmt.Printf("[FormHandler] Database insertion error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user data"})
		return
	}

	fmt.Printf("[FormHandler] Form data successfully saved for user: %s\n", email)
	c.Redirect(http.StatusFound, "/dashboard")
}

func FormUserDataHandler(c *gin.Context) {
	fmt.Println("[FormUserDataHandler] Starting data retrieval process...")

	email := c.Query("email")
	if email == "" {
		fmt.Println("[FormUserDataHandler] Error: No email provided in query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	fmt.Printf("[FormUserDataHandler] Fetching data for email: %s\n", email)

	// Query user details
	var user UserDetails
	query := `
		SELECT name, gender, age, activity_level, goals, height, weight, 
			   target_weight, diseases, email, diet_plan, healthscore, track 
		FROM user_details WHERE email = $1`

	err := DB.DB.QueryRow(query, email).Scan(
		&user.Name, &user.Gender, &user.Age, &user.ActivityLevel,
		&user.Goals, &user.Height, &user.Weight, &user.TargetWeight,
		&user.Diseases, &user.Email, &user.DietPlan, &user.Healthscore, &user.Track)

	if err != nil {
		fmt.Printf("[FormUserDataHandler] Database query error: %v\n", err)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	fmt.Printf("[FormUserDataHandler] Raw user data from DB: %+v\n", user)

	// Calculate BMI
	heightInMeters := user.Height / 100.0
	bmi := user.Weight / (heightInMeters * heightInMeters)
	bmi = math.Round(bmi*100) / 100 // Round to 2 decimal places

	// Calculate Health Score based on various factors
	healthScore := calculateHealthScore(user, bmi)

	// Update health score in database if it has changed
	if healthScore != user.Healthscore {
		_, err = DB.DB.Exec("UPDATE user_details SET healthscore = $1 WHERE email = $2", healthScore, email)
		if err != nil {
			fmt.Printf("[FormUserDataHandler] Failed to update health score: %v\n", err)
		}
		user.Healthscore = healthScore
	}

	response := gin.H{
		"name":           user.Name,
		"gender":         user.Gender,
		"age":            user.Age,
		"activity_level": user.ActivityLevel,
		"goals":          user.Goals,
		"height":         user.Height,
		"weight":         user.Weight,
		"target_weight":  user.TargetWeight,
		"diseases":       user.Diseases,
		"email":          user.Email,
		"diet_plan":      user.DietPlan.String,
		"healthscore":    healthScore,
		"track":          user.Track,
		"bmi":            bmi,
	}

	fmt.Printf("[FormUserDataHandler] Sending response: %+v\n", response)
	c.JSON(http.StatusOK, response)
}

// calculateHealthScore calculates a health score based on various health factors
func calculateHealthScore(user UserDetails, bmi float64) int {
	score := 100 // Start with max score

	// BMI factor (normal BMI range is 18.5-24.9)
	if bmi < 18.5 {
		score -= 10 // Underweight
	} else if bmi >= 25 && bmi < 30 {
		score -= 10 // Overweight
	} else if bmi >= 30 {
		score -= 20 // Obese
	}

	// Age factor
	if user.Age > 60 {
		score -= 10
	} else if user.Age < 18 {
		score -= 5
	}

	// Activity level factor
	switch strings.ToLower(user.ActivityLevel) {
	case "little":
		score -= 15
	case "moderate":
		score -= 5
	case "active":
		// No deduction for active lifestyle
	}

	// Medical conditions factor
	diseases := strings.Split(user.Diseases, ",")
	score -= len(diseases) * 5 // Deduct 5 points for each medical condition

	// Weight vs Target weight factor
	weightDiff := math.Abs(user.Weight - user.TargetWeight)
	if weightDiff > 20 {
		score -= 15
	} else if weightDiff > 10 {
		score -= 10
	} else if weightDiff > 5 {
		score -= 5
	}

	// Ensure score stays within 0-100 range
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	// Convert to 0-10 scale
	return score / 10
}

func AppendMealsHandler(c *gin.Context) {
	fmt.Println("[AppendMealsHandler] Parsing multipart form")
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println("[AppendMealsHandler] Error parsing form:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("[AppendMealsHandler] Received files:", form.File)

	var wg sync.WaitGroup
	wg.Add(len(form.File))
	processedImages := make(chan map[string]interface{}, len(form.File))

	for fieldName, files := range form.File {
		for _, file := range files {
			go func(fieldName string, file *multipart.FileHeader) {
				defer wg.Done()
				ImageProcess(c, fieldName, file, processedImages)
			}(fieldName, file)
		}
	}

	wg.Wait()
	close(processedImages)

	var breakfastResult map[string]interface{}
	var lunchResult map[string]interface{}
	var dinnerResult map[string]interface{}

	for processedImage := range processedImages {
		fieldName := processedImage["fieldName"].(string)
		fmt.Println("[AppendMealsHandler] Processed image for:", fieldName)
		if fieldName == "breakfast_img" {
			breakfastResult = processedImage["data"].(map[string]interface{})
		} else if fieldName == "lunch_img" {
			lunchResult = processedImage["data"].(map[string]interface{})
		} else if fieldName == "dinner_img" {
			dinnerResult = processedImage["data"].(map[string]interface{})
		}
	}

	email := c.PostForm("email")
	date := c.PostForm("date")
	breakfast := c.PostForm("breakfast")
	lunch := c.PostForm("lunch")
	dinner := c.PostForm("dinner")
	weight := c.PostForm("weight")

	formData := make(map[string]interface{})
	formData["email"] = email
	formData["date"] = date

	if breakfast == "" && breakfastResult != nil {
		formData["breakfast"] = breakfastResult
	} else if breakfast != "" && breakfastResult == nil {
		formData["breakfast"] = breakfast
	} else {
		fmt.Println("[AppendMealsHandler] Breakfast error")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Breakfast value empty."})
		return
	}

	if lunchResult != nil {
		formData["lunch"] = lunchResult
	} else if lunch != "" {
		formData["lunch"] = lunch
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lunch missing"})
		return
	}

	if dinner == "" && dinnerResult != nil {
		formData["dinner"] = dinnerResult
	} else if lunch != "" && dinnerResult == nil {
		formData["dinner"] = dinner
	} else {
		fmt.Println("[AppendMealsHandler] Dinner error")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dinner value empty."})
		return
	}

	formData["weight"] = weight
	fmt.Println("[AppendMealsHandler] Final form data to save:", formData)

	err = DB.AppendMeals(formData)
	if err != nil {
		fmt.Println("[AppendMealsHandler] Couldn't track calories:", err)
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "message sent"})
}

func ImageProcess(c *gin.Context, fieldName string, file *multipart.FileHeader, processedImages chan map[string]interface{}) {
	fmt.Println("[ImageProcess] Processing image:", file.Filename)

	openedFile, err := file.Open()
	if err != nil {
		fmt.Println("[ImageProcess] Failed to open file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer openedFile.Close()

	imageData, err := ioutil.ReadAll(openedFile)
	if err != nil {
		fmt.Println("[ImageProcess] Failed to read file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	imageType := http.DetectContentType(imageData)
	switch imageType {
	case "image/png":
		imageType = "png"
	case "image/webp":
		imageType = "webp"
	case "image/jpeg":
		imageType = "png"
	default:
		imageType = "png"
	}
	fmt.Println("[ImageProcess] Image type detected:", imageType)

	result_map, err := AI.AnalyzeImageWithGroq(imageData, imageType)
	if err != nil {
		fmt.Println("[ImageProcess] Error in AI.AnalyzeImageWithGroq:", err)
	}

	processedImages <- map[string]interface{}{
		"fieldName": fieldName,
		"data":      result_map,
	}
}

func GetMime(header *multipart.FileHeader) string {
	contentType := header.Header.Get("Content-Type")
	mime, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		fmt.Println("[GetMime] Error parsing content type:", err)
		return ""
	}

	fmt.Println("[GetMime] MIME type:", mime)
	return mime
}

func UpdateDietHandler(c *gin.Context) {
	email := c.PostForm("email")
	diet := c.PostForm("diet_plan")
	healthscore := c.PostForm("healthscore")
	fmt.Println("[UpdateDietHandler] Email:", email, "Healthscore:", healthscore)

	hs, err := strconv.Atoi(healthscore)
	if err != nil {
		fmt.Println("[UpdateDietHandler] Error converting healthscore:", err)
		hs = 0
	}

	err = DB.UpdateDiet(email, diet, hs)
	if err != nil {
		fmt.Println("[UpdateDietHandler] Error updating diet:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "couldn't get updated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "plan updated"})
}

func GenDietPlan(c *gin.Context) {
	fmt.Println("[GenDietPlan] Starting diet plan generation...")

	var userData map[string]interface{}
	if err := c.ShouldBindJSON(&userData); err != nil {
		fmt.Println("[GenDietPlan] JSON Bind Error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the received data
	fmt.Printf("[GenDietPlan] Received user data: %+v\n", userData)

	// Extract and validate email
	emailVal, ok := userData["email"].(string)
	if !ok || emailVal == "" {
		fmt.Printf("[GenDietPlan] Email is missing or invalid: %v\n", userData["email"])
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is missing or invalid"})
		return
	}

	// Create a prompt for the AI model
	prompt := fmt.Sprintf(
		"Generate a detailed diet plan for:\nName: %v\nAge: %v\nGender: %v\n"+
			"Activity Level: %v\nGoals: %v\nHeight: %v cm\nWeight: %v kg\n"+
			"Target Weight: %v kg\nMedical Conditions: %v\n\n"+
			"Please provide a detailed daily diet plan including breakfast, lunch, dinner, and snacks. "+
			"Include portion sizes and timing. Consider their medical conditions and fitness goals.",
		userData["name"], userData["age"], userData["gender"],
		userData["activityLevel"], userData["goals"], userData["height"],
		userData["weight"], userData["tweight"], userData["disease"])

	fmt.Println("[GenDietPlan] Sending prompt to AI:", prompt)

	// Generate diet plan via Groq AI
	dietPlan, err := AI.GenGroqAI(prompt)
	if err != nil {
		fmt.Println("[GenDietPlan] Failed to generate diet plan:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't generate diet plan"})
		return
	}
	fmt.Println("✅ Diet plan generated successfully")

	// Store result in DB
	err = DB.UpdateDiet(emailVal, dietPlan, 0)
	if err != nil {
		fmt.Println("[GenDietPlan] DB update error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't store diet plan in database"})
		return
	}
	fmt.Println("✅ Diet plan saved to database")

	c.JSON(http.StatusOK, gin.H{"diet_plan": dietPlan})
}

// GetUserBasicInfo returns just the name and email of the user
func GetUserBasicInfo(c *gin.Context) {
	fmt.Println("[GetUserBasicInfo] Starting basic info retrieval...")

	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	var name string
	err := DB.DB.QueryRow("SELECT name FROM user_details WHERE email = $1", email).Scan(&name)
	if err != nil {
		fmt.Printf("[GetUserBasicInfo] Error fetching name: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":  name,
		"email": email,
	})
}

// GetUserBMI calculates and returns the user's BMI
func GetUserBMI(c *gin.Context) {
	fmt.Println("[GetUserBMI] Starting BMI calculation...")

	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	var height, weight float64
	err := DB.DB.QueryRow("SELECT height, weight FROM user_details WHERE email = $1", email).Scan(&height, &weight)
	if err != nil {
		fmt.Printf("[GetUserBMI] Error fetching metrics: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user metrics"})
		return
	}

	heightInMeters := height / 100.0
	bmi := weight / (heightInMeters * heightInMeters)
	bmi = math.Round(bmi*100) / 100 // Round to 2 decimal places

	c.JSON(http.StatusOK, gin.H{
		"bmi": bmi,
	})
}

// GetUserHealthScore returns the user's health score
func GetUserHealthScore(c *gin.Context) {
	fmt.Println("[GetUserHealthScore] Starting health score retrieval...")

	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	var healthscore int
	err := DB.DB.QueryRow("SELECT healthscore FROM user_details WHERE email = $1", email).Scan(&healthscore)
	if err != nil {
		fmt.Printf("[GetUserHealthScore] Error fetching health score: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch health score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"healthscore": healthscore,
	})
}
