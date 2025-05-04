import { initializeApp } from "https://www.gstatic.com/firebasejs/10.10.0/firebase-app.js";
import { getAuth, signOut, onAuthStateChanged } from "https://www.gstatic.com/firebasejs/10.10.0/firebase-auth.js";

const firebaseConfig = {
    apiKey: "AIzaSyAOPZjFuX_9WWuVwso8RtEnY6W4ntt4ZBg",
    authDomain: "blissful-bites-9985b.firebaseapp.com",
    projectId: "blissful-bites-9985b",
    storageBucket: "blissful-bites-9985b.appspot.com",
    messagingSenderId: "1070336763456",
    appId: "1:1070336763456:web:d4eefe70dd2d4c53fd61a5",
    measurementId: "G-LWJW5JXMEL"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const auth = getAuth(app);

auth.onAuthStateChanged(console.log)
var email;
onAuthStateChanged(auth, (user) => {
    if (user) {
        // User is signed in
        email = user.email
        console.log("User is signed in:", user);
        console.log(email);

    } else {
        // No user is signed in
        console.log("No user is signed in");
        // You can handle this case, for example, redirecting the user to the login page
    }
});

const logout = document.getElementById("logoutButton");
logout.addEventListener('click', () => {
    signOut(auth).then(() => {
        // Sign-out successful
        console.log("User signed out successfully");
        window.location.href = "/login";

    }).catch((error) => {
        console.error("Error signing out:", error);
    });
});

const logout2 = document.getElementById("logoutButton2");
logout2.addEventListener('click', () => {
    signOut(auth).then(() => {
        // Sign-out successful
        console.log("User signed out successfully");
        window.location.href = "/login";

    }).catch((error) => {
        console.error("Error signing out:", error);
    });
});