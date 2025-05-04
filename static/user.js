
const urlParams = new URLSearchParams(window.location.search);
const email = urlParams.get('email');
console.log(email)
document.getElementById("email").value = email;
fetch(`/userDetails?email=${email}`)
    .then(response => response.json())
    .then(user => {
        // Display user details on the page
        const name = document.getElementById('name');
        name.innerHTML = `${user.name}`;
        const age = document.getElementById('age');
        age.innerHTML = `${user.age}`;
        const gender = document.getElementById('gender');
        gender.innerHTML = `${user.gender}`;
        const al = document.getElementById('al');
        al.innerHTML = `${user.activity_level}`;
        const height = document.getElementById('height');
        height.innerHTML = `${user.height}`;
        const weight = document.getElementById('weight');
        weight.innerHTML = `${user.weight}`;
        const tweight = document.getElementById('tweight');
        tweight.innerHTML = `${user.target_weight}`;
        const goals = document.getElementById('goals');
        goals.innerHTML = `${user.goals}`;
        const diseases = document.getElementById('diseases');
        diseases.innerHTML = `${user.diseases}`;
        const curr_hs = document.getElementById('curr_hs');
        curr_hs.innerHTML = `${user.healthscore}`;
        const curr_diet = document.getElementById('curr_diet');
        curr_diet.innerHTML = `${user.diet_plan}`;
        
        
        
        // const userDetailsDiv = document.getElementById('userDetails');
    })
    .catch(error => {
        console.error('Error fetching user details:', error);
    });

document.getElementById('updateDietForm').addEventListener('submit', function (event) {
    event.preventDefault(); // Prevent the form from submitting normally

    // Get form data
    var formData = new FormData(this);
    console.log(formData)
    // Send form data to server
    fetch('/updateDiet', {
        method: 'POST',
        body: formData
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            // Clear the form input
            document.getElementById('diet_plan').value = '';
            document.getElementById('healthscore').value = '';

            // Display success message
            alert('Diet Plan updated successfully!');
        })
        .catch(error => {
            // Display error message
            alert('Oops! Something went wrong. Please try again later.');
        });
});