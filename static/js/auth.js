const API_BASE = "api/user"; // Adjust this if needed

// Temporarily store registration data
let registrationData = {};

// Handle registration step 1 (email and password)
document.getElementById("registerStep1")?.addEventListener("submit", async function (e) {
    e.preventDefault();
    const formData = new FormData(this);
    const email = formData.get("email");
    const password = formData.get("password");

    if (!email || !password) {
        alert("Email and Password are required.");
        return;
    }

    // Save email and password for the next step
    registrationData.email = email;
    registrationData.password = password;

    alert("Step 1 completed. Proceeding to phone verification.");
    document.getElementById("step1").style.display = "none";
    document.getElementById("step2").style.display = "block";
});

// Handle registration step 2 (phone number and OTP verification)
document.getElementById("registerStep2")?.addEventListener("submit", async function (e) {
    e.preventDefault();
    const formData = new FormData(this);
    const phoneNumber = formData.get("phone_number");
    const countryCode = formData.get("country_code");

    if (!phoneNumber || !countryCode) {
        alert("Phone number and country code are required.");
        return;
    }

    try {
        // Request OTP
        const otpFormData = new FormData();
        intPhoneNumber = parseInt(phoneNumber);
        console.log(typeof(intPhoneNumber));
        otpFormData.append("phone_number", intPhoneNumber);
        for (const pair of otpFormData.entries()) {
            console.log(`${pair[0]}: ${pair[1]}`);
        }

        const otpResponse = await fetch(`http://localhost:8080/api/user/generate-otp`, {
            method: "POST",
            body: otpFormData,
        });

        const otpResult = await otpResponse.json();
        if (!otpResponse.ok) {
            alert(otpResult.message || "Failed to generate OTP");
            return;
        }

        const userOTP = prompt("Enter the OTP sent to your phone:");
        if (!userOTP) {
            alert("OTP is required to proceed.");
            return;
        }

        // Verify OTP
        const verifyFormData = new FormData();
        verifyFormData.append("phone_number", phoneNumber);
        verifyFormData.append("otp", userOTP);

        const verifyResponse = await fetch(`http://localhost:8080/api/user/verify-otp`, {
            method: "POST",
            body: verifyFormData,
        });

        const verifyResult = await verifyResponse.json();
        if (!verifyResponse.ok) {
            alert(verifyResult.message || "OTP verification failed");
            return;
        }

        // Save phone details for final registration
        registrationData.phone_number = phoneNumber;
        registrationData.country_code = countryCode;

        // Finalize registration
        const registerFormData = new FormData();
        registerFormData.append("email", registrationData.email);
        registerFormData.append("password", registrationData.password);
        registerFormData.append("phone_number", registrationData.phone_number);
        registerFormData.append("country_code", registrationData.country_code);

        const registerResponse = await fetch(`http://localhost:8080/api/user/register`, {
            method: "POST",
            body: registerFormData,
        });

        const registerResult = await registerResponse.json();
        if (!registerResponse.ok) {
            alert(registerResult.message || "Registration failed");
            return;
        }

        alert("Registration successful! Redirecting to login.");
        window.location.href = "../index.html"; // Redirect to login
    } catch (error) {
        console.error("Error during registration:", error);
        alert("An error occurred. Please try again.");
    }
});


// Handle login
document.getElementById("loginForm")?.addEventListener("submit", async function (e) {
    e.preventDefault();
    const formData = new FormData(this);
    const email = formData.get("email");
    const password = formData.get("password");
    console.log(formData);

    if (!email || !password) {
        alert("Email and Password are required.");
        return;
    }

    try {
        const response = await fetch(`http://localhost:8080/api/user/login`, {
            method: "POST",
            body: formData,
        });

        const result = await response.json();
        if (!response.ok) {
            alert(result.message || "Login failed");
            return;
        }

        localStorage.setItem("token", result.token); // Save JWT token
        console.log(result.token);
        alert("Login successful");
        window.location.href = "../html/profile.html"; // Redirect to profile
    } catch (error) {
        console.error("Error during login:", error);
        alert("An error occurred. Please try again.");
    }
});
