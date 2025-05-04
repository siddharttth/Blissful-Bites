package DB

import (
	"database/sql"
	// "encoding/json"
	"encoding/json"
	"fmt"

	// _ "github.com/lib/pq"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	// "reflect"
)

// const (
// 	host     = "localhost"
// 	port     = 5432
// 	user     = "postgres"
// 	password = "postgres"
// 	dbname   = "blissful_bites"
// )

var DB *sql.DB
var err error

func ConnectPsql(user string, pass string, host string, port string, dbname string) error {

	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, dbname)

	DB, err = sql.Open("pgx", psqlInfo)
	if err != nil {
		fmt.Println("[DB] Error connecting to postgres server.")
		return err
	}

	err = DB.Ping()

	return err

}

func DropTableUserDetails() error {
	_, err = DB.Exec(`DROP TABLE IF EXISTS user_details;`)
	if err != nil {
		fmt.Println("[DROP UD] Cant drop.", err.Error())
		return err
	}
	fmt.Println("✅ Dropped Table UserDetails")
	return nil
}

func CreateTableUserCredentials() error {
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS user_credentials (
		email VARCHAR(100) PRIMARY KEY,
		password VARCHAR(255) NOT NULL
	);
	`)
	if err != nil {
		fmt.Println("[CREATE UC] Cant create.", err.Error())
		return err
	}
	fmt.Println("✅ Created Table UserCredentials")
	return nil
}

func CreateTableUserDetails() error {
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS user_details (
		name VARCHAR(100) NOT NULL,
		gender VARCHAR(10) NOT NULL,
		age INTEGER NOT NULL,
		activity_level VARCHAR(20) NOT NULL,
		goals TEXT NOT NULL,
		height FLOAT NOT NULL,
		weight FLOAT NOT NULL,
		target_weight FLOAT NOT NULL,
		diseases TEXT NOT NULL,
		email VARCHAR(100) PRIMARY KEY,
		diet_plan TEXT,
		healthscore INTEGER NOT NULL,
		track JSONB,
		dm TEXT
	);	
	
	`)
	if err != nil {
		fmt.Println("[CREATE UD] Cant create.", err.Error())
		return err
	}
	fmt.Println("✅ Created Table UserDetails")
	return nil
}

func InsertUserData(values map[string]interface{}) error {
	fmt.Println("[DB] Starting user data insertion...")
	fmt.Printf("[DB] Received values: %+v\n", values)

	// Extract and validate email first
	email, ok := values["email"].(string)
	if !ok || email == "" {
		return fmt.Errorf("invalid or missing email")
	}
	fmt.Printf("[DB] Processing data for email: %s\n", email)

	// Extract and validate other fields
	name, _ := values["name"].(string)
	gender, _ := values["gender"].(string)
	ageStr, _ := values["age"].(string)
	age, err := strconv.Atoi(ageStr)
	if err != nil {
		fmt.Printf("[DB] Age conversion error: %v\n", err)
		return fmt.Errorf("invalid age format: %w", err)
	}

	activityLevel, _ := values["activityLevel"].(string)

	heightStr, _ := values["height"].(string)
	height, err := strconv.ParseFloat(heightStr, 64)
	if err != nil {
		fmt.Printf("[DB] Height conversion error: %v\n", err)
		return fmt.Errorf("invalid height format: %w", err)
	}

	weightStr, _ := values["weight"].(string)
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		fmt.Printf("[DB] Weight conversion error: %v\n", err)
		return fmt.Errorf("invalid weight format: %w", err)
	}

	targetWeightStr, _ := values["tweight"].(string)
	targetWeight, err := strconv.ParseFloat(targetWeightStr, 64)
	if err != nil {
		fmt.Printf("[DB] Target weight conversion error: %v\n", err)
		return fmt.Errorf("invalid target weight format: %w", err)
	}

	disease := ""
	goal := ""

	goals, ok := values["goals"].([]string)
	if !ok {
		goal, _ = values["goals"].(string)
	}

	diseases, ok := values["disease"].([]string)
	if !ok {
		disease, _ = values["disease"].(string)
	}

	hsStr, _ := values["healthscore"].(string)
	hs, err := strconv.Atoi(hsStr)
	if err != nil {
		fmt.Printf("[DB] Healthscore conversion error: %v\n", err)
		return fmt.Errorf("invalid healthscore format: %w", err)
	}

	// Convert arrays to comma-separated strings
	var goalsStr string
	var diseasesStr string
	if goal == "" {
		goalsStr = strings.Join(goals, ",")
	} else {
		goalsStr = goal
	}
	if disease == "" {
		diseasesStr = strings.Join(diseases, ",")
	} else {
		diseasesStr = disease
	}

	// Use UPSERT (INSERT ... ON CONFLICT DO UPDATE) to handle both new and existing records
	query := `
        INSERT INTO user_details (
            email, name, gender, age, activity_level, goals, 
            height, weight, target_weight, diseases, healthscore
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (email) DO UPDATE SET
            name = EXCLUDED.name,
            gender = EXCLUDED.gender,
            age = EXCLUDED.age,
            activity_level = EXCLUDED.activity_level,
            goals = EXCLUDED.goals,
            height = EXCLUDED.height,
            weight = EXCLUDED.weight,
            target_weight = EXCLUDED.target_weight,
            diseases = EXCLUDED.diseases,
            healthscore = EXCLUDED.healthscore
        RETURNING email;
    `

	var returnedEmail string
	err = DB.QueryRow(query,
		email, name, gender, age, activityLevel, goalsStr,
		height, weight, targetWeight, diseasesStr, hs,
	).Scan(&returnedEmail)

	if err != nil {
		fmt.Printf("[DB] Database error: %v\n", err)
		return fmt.Errorf("failed to insert/update user data: %w", err)
	}

	fmt.Printf("[DB] Successfully saved data for user: %s\n", returnedEmail)
	return nil
}

func ReadTrack(email string) ([]byte, error) {

	query := "SELECT track FROM user_details WHERE email = $1"
	var jsonData []byte
	err := DB.QueryRow(query, email).Scan(&jsonData)
	if err != nil {
		fmt.Println("[Reading Track]", err)
		return nil, err
	}
	return jsonData, nil
}

func CheckEmailExists(email string) (bool, error) {

	query := "SELECT COUNT(*) FROM user_details WHERE email = $1"
	var count int
	err := DB.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to execute query: %w", err)
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func FetchExistingJSONData(email string) ([]map[string]interface{}, error) {

	var track map[string]interface{}
	var trackData []map[string]interface{}
	var trackDataString sql.NullString

	query := "SELECT track FROM user_details WHERE email = $1"
	err := DB.QueryRow(query, email).Scan(&trackDataString)
	if err == fmt.Errorf("sql: no rows in result set") {
		return []map[string]interface{}{}, nil
	}
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if trackDataString.Valid {
		err = json.Unmarshal([]byte(trackDataString.String), &trackData)
		if err != nil {
			err2 := json.Unmarshal([]byte(trackDataString.String), &track)
			if err2 != nil {
				fmt.Println(err2)
				return nil, fmt.Errorf("error parsing JSONB data: %v", err2)
			}
			trackData = []map[string]interface{}{track}
			return trackData, nil
		}
	} else {
		trackData = []map[string]interface{}{}
	}
	return trackData, nil
}

func AppendMeals(values map[string]interface{}) error {

	email, _ := values["email"].(string)
	delete(values, "email")

	existingMap, err := FetchExistingJSONData(email)
	if err != nil {
		return fmt.Errorf("failed to fetch existing JSON data: %w", err)
	}

	existingMap = append(existingMap, values)

	newData, err := json.Marshal(existingMap)
	if err != nil {
		return fmt.Errorf("failed to marshal merged data to JSON: %w", err)
	}

	query := `
        UPDATE user_details
        SET track = $1 
		WHERE email = $2
    `

	_, err = DB.Exec(query, newData, email)
	if err != nil {
		return fmt.Errorf("failed to update JSON data: %w", err)
	}

	return nil
}

func ReadAllUsers(query string) ([]map[string]interface{}, error) {
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize a slice to hold the results
	results := []map[string]interface{}{}

	// Iterate over the rows
	for rows.Next() {
		// Create a map to hold the values of the current row
		rowData := make(map[string]interface{})

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		// Create a slice to store values for Scan
		values := make([]interface{}, len(columns))
		for i := range columns {
			values[i] = new(interface{})
		}

		// Scan the values of the current row into the map
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}

		// Convert interface{} values to appropriate types and store them in the map
		for i, col := range columns {
			rowData[col] = *(values[i].(*interface{}))
		}

		// Append the map to the results slice
		results = append(results, rowData)
	}

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func ReadRowData(primaryKeyValue string) (map[string]interface{}, error) {
	rowData := make(map[string]interface{})

	// Build SQL query
	query := fmt.Sprintf("SELECT * FROM user_details WHERE email = $1")

	// Execute the query and get the rows object
	rows, err := DB.Query(query, primaryKeyValue)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting column names: %v", err)
	}

	// Create a slice of interface{} to hold column values
	values := make([]interface{}, len(columns))

	// Create pointers to each element in the values slice
	for i := range values {
		values[i] = new(interface{})
	}

	// Iterate over rows
	for rows.Next() {
		// Scan row data into the slice
		err = rows.Scan(values...)
		if err != nil {
			return nil, fmt.Errorf("error scanning row data: %v", err)
		}

		// Add data to the map
		for i, colName := range columns {
			rowData[colName] = *(values[i].(*interface{}))
		}
	}

	return rowData, nil
}

func UpdateDM(email string, message string) error {

	stmt, err := DB.Prepare("UPDATE user_details SET dm = $1 WHERE email = $2")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(message, email)
	if err != nil {
		return err
	}

	fmt.Println("Text field updated successfully!")
	return nil
}

func UpdateDiet(email string, diet string, healthscore int) error {
	var query string
	if healthscore == 0 {
		query = "UPDATE user_details SET diet_plan = $1 WHERE email = $2"
	} else {
		query = "UPDATE user_details SET diet_plan = $1, healthscore = $2 WHERE email = $3"
	}
	stmt, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var err2 error
	if healthscore == 0 {
		_, err2 = stmt.Exec(diet, email)
	} else {
		_, err2 = stmt.Exec(diet, healthscore, email)
	}
	if err2 != nil {
		return err2
	}

	fmt.Println("Diet updated successfully!")
	return nil
}

func MigrateDB() error {
	fmt.Println("[DB] Starting database migrations...")

	// Create user_credentials table
	_, err := DB.Exec(`
	CREATE TABLE IF NOT EXISTS user_credentials (
		email VARCHAR(100) PRIMARY KEY,
		password VARCHAR(255) NOT NULL
	);`)
	if err != nil {
		return fmt.Errorf("failed to create user_credentials table: %w", err)
	}
	fmt.Println("✅ user_credentials table ready")

	// Create user_details table with email as PRIMARY KEY
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS user_details (
		email VARCHAR(100) PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		gender VARCHAR(10) NOT NULL,
		age INTEGER NOT NULL,
		activity_level VARCHAR(20) NOT NULL,
		goals TEXT NOT NULL,
		height FLOAT NOT NULL,
		weight FLOAT NOT NULL,
		target_weight FLOAT NOT NULL,
		diseases TEXT NOT NULL,
		diet_plan TEXT,
		healthscore INTEGER NOT NULL,
		track JSONB,
		dm TEXT,
		FOREIGN KEY (email) REFERENCES user_credentials(email)
	);`)
	if err != nil {
		return fmt.Errorf("failed to create user_details table: %w", err)
	}
	fmt.Println("✅ user_details table ready")

	fmt.Println("✅ All database migrations completed successfully")
	return nil
}
