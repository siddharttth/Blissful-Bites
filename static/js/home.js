// Remove Firebase imports and config
document.addEventListener('DOMContentLoaded', async () => {
    console.log("[Init] Starting dashboard initialization...");

    // Check for session email
    const userEmail = sessionStorage.getItem('userEmail');
    console.log("[Auth] Checking session email:", userEmail);

    if (!userEmail) {
        console.log("[Auth] No user session found, redirecting to login");
        window.location.href = "/login";
        return;
    }

    console.log("[Auth] Session found for:", userEmail);

    try {
        // Initial data fetch
        console.log("[Init] Fetching initial user data");
        await fetchUserDetails(userEmail);
        console.log("[Init] Initial data fetch complete");

        // Set up diet plan button click handler
        const genDietPlanButton = document.getElementById('gendietplan');
        if (genDietPlanButton) {
            console.log("[UI] Setting up diet plan button handler");
            genDietPlanButton.onclick = async function() {
                console.log("[Event] Diet plan button clicked");
                const dietPlanElement = document.getElementById('diet_plan');
                
                try {
                    // Disable button and show loading state
                    this.disabled = true;
                    console.log("[UI] Button disabled");
                    
                    if (dietPlanElement) {
                        dietPlanElement.textContent = 'Generating your personalized diet plan...';
                        console.log("[UI] Updated to loading state");
                    }

                    // Get email from session storage
                    const userEmail = sessionStorage.getItem('userEmail');
                    if (!userEmail) {
                        throw new Error('No user email found in session');
                    }

                    // Fetch fresh user data
                    console.log("[API] Fetching fresh user data for email:", userEmail);
                    const response = await fetch(`/userDetails?email=${userEmail}`);
                    if (!response.ok) {
                        throw new Error(`Failed to fetch user data: ${response.status}`);
                    }
                    const userData = await response.json();
                    console.log("[API] Fresh user data received:", userData);

                    // Prepare data for diet plan generation
                    const dietPlanData = {
                        email: userEmail,
                        name: userData.name,
                        age: userData.age,
                        gender: userData.gender,
                        activityLevel: userData.activity_level,
                        goals: userData.goals,
                        height: userData.height,
                        weight: userData.weight,
                        tweight: userData.target_weight,
                        disease: userData.diseases,
                        healthscore: userData.healthscore,
                        track: userData.track
                    };
                    console.log("[API] Prepared diet plan request data:", dietPlanData);

                    // Call diet plan generation API
                    console.log("[API] Calling diet plan generation endpoint");
                    const dietResponse = await fetch('/genDietPlan', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify(dietPlanData)
                    });

                    if (!dietResponse.ok) {
                        throw new Error(`Diet plan generation failed: ${dietResponse.status}`);
                    }

                    const dietData = await dietResponse.json();
                    console.log("[API] Diet plan received:", dietData);

                    // Update UI with diet plan
                    if (dietPlanElement && dietData.diet_plan) {
                        if (typeof dietData.diet_plan === 'string') {
                            // Clean up and format the diet plan text
                            const cleanDietPlan = formatDietPlan(dietData.diet_plan);
                            dietPlanElement.innerHTML = cleanDietPlan;
                        } else if (Array.isArray(dietData.diet_plan) || typeof dietData.diet_plan === 'object') {
                            dietPlanElement.innerHTML = JSON.stringify(dietData.diet_plan, null, 2);
                        } else {
                            throw new Error('Invalid diet plan data type received');
                        }
                        console.log("[UI] Diet plan displayed successfully");
                    } else {
                        throw new Error('Invalid diet plan data received');
                    }
                } catch (error) {
                    console.error('[Error] Diet plan generation failed:', error);
                    if (dietPlanElement) {
                        dietPlanElement.textContent = 'Failed to generate diet plan. Please try again.';
                        console.log("[UI] Error message displayed");
                    }
                } finally {
                    // Re-enable button
                    this.disabled = false;
                    console.log("[UI] Button re-enabled");
                }
            };
            console.log("[UI] Diet plan button handler setup complete");
        } else {
            console.error("[Error] Diet plan button not found in DOM");
        }
    } catch (error) {
        console.error("[Error] Dashboard initialization failed:", error);
    }
});

async function fetchUserDetails(userEmail) {
    console.log("[API] Starting user details fetch for:", userEmail);
    
    if (!userEmail) {
        throw new Error("No email provided for fetching user details");
    }

    try {
        const url = `/userDetails?email=${userEmail}`;
        console.log("[API] Fetching from URL:", url);

        const response = await fetch(url);
        console.log("[API] Response status:", response.status);

        if (!response.ok) {
            throw new Error(`Network response was not ok: ${response.status}`);
        }

        const data = await response.json();
        console.log("[API] User data received:", data);

        // Calculate BMI
        const heightInM = data.height / 100;
        const bmi = (data.weight / (heightInM * heightInM)).toFixed(2);
        console.log("[Calc] Calculated BMI:", bmi);

        // Update UI elements
        console.log("[UI] Updating dashboard elements");
        const elements = {
            name: document.getElementById('name'),
            hs: document.getElementById('hs'),
            bmi: document.getElementById('bmi'),
            dietPlan: document.getElementById('diet_plan')
        };

        // Log found elements
        console.log("[UI] Found elements:", Object.keys(elements).filter(k => elements[k]));

        // Update each element with null checks
        if (elements.name) {
            elements.name.textContent = data.name || 'User';
            console.log("[UI] Name updated:", elements.name.textContent);
        }
        if (elements.hs) {
            elements.hs.textContent = data.healthscore || '0';
            console.log("[UI] Health score updated:", elements.hs.textContent);
        }
        if (elements.bmi) {
            elements.bmi.textContent = bmi || '0';
            console.log("[UI] BMI updated:", elements.bmi.textContent);
        }
        if (elements.dietPlan) {
            elements.dietPlan.textContent = data.diet_plan || 'Click Generate Diet Plan to get your personalized diet plan';
            console.log("[UI] Diet plan placeholder set");
        }

        // Process track data if available
        if (data.track && data.track.length > 0) {
            console.log("[Charts] Processing track data");
            const weightMap = extractWeightValues(data.track);
            const calorieMap = extractCalorieValues(data.track);

            createWeightChart(weightMap, "weightChart", "weight");
            createWeightChart(calorieMap, "calorieChart", "Total Calories");
            console.log("[Charts] Charts created successfully");
        } else {
            console.log("[Charts] No track data available");
        }

        return data;
    } catch (error) {
        console.error("[Error] Failed to fetch user details:", error);
        throw error;
    }
}

// Chart-related functions
function extractWeightValues(trackData) {
    console.log("[Charts] Starting weight data extraction");
    const weightMap = {};
    
    if (trackData && trackData.length > 0) {
        trackData.forEach(function (track) {
            if (track.hasOwnProperty("weight") && track.hasOwnProperty("date")) {
                if (weightMap.hasOwnProperty(track.date)) {
                    weightMap[track.date].push(track.weight);
                } else {
                    weightMap[track.date] = [track.weight];
                }
            }
        });
        console.log("[Charts] Extracted weight data points:", Object.keys(weightMap).length);
    } else {
        console.log("[Charts] No track data available for weight extraction");
    }
    
    return weightMap;
}

function extractCalorieValues(trackData) {
    console.log("[Charts] Starting calorie data extraction");
    const calorieMap = {};
    
    if (trackData && trackData.length > 0) {
        trackData.forEach(function (track) {
            if (track.hasOwnProperty("date")) {
                let totalCalories = 0;
                if (track.hasOwnProperty("breakfast") && track.breakfast.hasOwnProperty("Total calories")) {
                    totalCalories += track.breakfast["Total calories"];
                }
                if (track.hasOwnProperty("lunch") && track.lunch.hasOwnProperty("Total calories")) {
                    totalCalories += track.lunch["Total calories"];
                }
                if (track.hasOwnProperty("dinner") && track.dinner.hasOwnProperty("Total calories")) {
                    totalCalories += track.dinner["Total calories"];
                }
                calorieMap[track.date] = totalCalories;
            }
        });
        console.log("[Charts] Extracted calorie data points:", Object.keys(calorieMap).length);
    } else {
        console.log("[Charts] No track data available for calorie extraction");
    }
    
    return calorieMap;
}

// Function to format the diet plan text
function formatDietPlan(dietPlan) {
    // Remove excessive stars and clean up formatting
    let formattedPlan = dietPlan
        .replace(/\*\*/g, '') // Remove double stars
        .replace(/\*/g, '')   // Remove single stars
        .replace(/\+/g, '<br>') // Replace + with line breaks
        .trim();

    // Split into sections
    const sections = formattedPlan.split('<br>').filter(section => section.trim());

    // Create HTML with proper formatting
    const html = sections.map(section => {
        section = section.trim();
        
        // Check if it's a main heading (contains "Diet Plan" or "Daily Calorie Target")
        if (section.includes('Diet Plan') || section.includes('Daily Calorie Target')) {
            return `<h3 class="text-xl font-bold mb-2">${section}</h3>`;
        }
        
        // Check if it's a subheading (contains a time)
        else if (section.includes('am)') || section.includes('pm)')) {
            return `<p class="font-semibold mb-2">${section}</p>`;
        }
        
        // Regular text
        else {
            return `<p class="mb-2">${section}</p>`;
        }
    }).join('');

    return `<div class="diet-plan p-4 bg-white rounded-lg shadow">
                ${html}
            </div>`;
}

// Logout handlers
const logout = document.getElementById("logoutButton");
if (logout) {
    logout.addEventListener('click', () => {
        console.log("[Auth] User clicked logout");
        sessionStorage.removeItem('userEmail');
        window.location.href = "/login";
    });
}

const logout2 = document.getElementById("logoutButton2");
if (logout2) {
    logout2.addEventListener('click', () => {
        console.log("[Auth] User clicked mobile logout");
        sessionStorage.removeItem('userEmail');
        window.location.href = "/login";
    });
}