const API_BASE = "localhost:8081"; // Adjust this if needed

// Handle registration step 1 (email and password)
document.getElementById("registerStep1")?.addEventListener("submit", async function (e) {
    e.preventDefault();
    const formData = new FormData(this);
    const email = formData.get("email");
    const password = formData.get("password");

    try {
        const response = await fetch(`${API_BASE}/register`, {
            method: "POST",
            body: new URLSearchParams({ email, password }),
        });

        const result = await response.json();
        if (!response.ok) {
            alert(result.message || "Registration failed at Step 1");
            return;
        }

        alert("Step 1 completed successfully. Proceeding to phone verification.");
        document.getElementById("step1").style.display = "none";
        document.getElementById("step2").style.display = "block";
    } catch (error) {
        console.log(error);
        console.error("Error during registration step 1:", error);
        alert("An error occurred. Please try again.");
    }
});

// Handle registration step 2 (phone number and OTP verification)
document.getElementById("registerStep2")?.addEventListener("submit", async function (e) {
    e.preventDefault();
    const formData = new FormData(this);
    const phoneNumber = formData.get("phone_number");
    const countryCode = formData.get("country_code");

    try {
        // Request OTP
        const otpResponse = await fetch(`${API_BASE}/generate-otp`, {
            method: "POST",
            body: new URLSearchParams({ phone_number: phoneNumber }),
        });

        const otpResult = await otpResponse.json();
        if (!otpResponse.ok) {
            alert(otpResult.message || "Failed to generate OTP");
            return;
        }

        const userOTP = prompt("Enter the OTP sent to your phone:");
        if (!userOTP) {
            alert("OTP is required to proceed");
            return;
        }

        // Verify OTP
        const verifyResponse = await fetch(`${API_BASE}/verify-otp`, {
            method: "POST",
            body: new URLSearchParams({ phone_number: phoneNumber, otp: userOTP }),
        });

        const verifyResult = await verifyResponse.json();
        if (!verifyResponse.ok) {
            alert(verifyResult.message || "OTP verification failed");
            return;
        }

        alert("Phone verification successful! Registration complete. Redirecting to login.");
        window.location.href = "../index.html"; // Redirect to login
    } catch (error) {
        console.error("Error during phone verification:", error);
        alert("An error occurred. Please try again.");
    }
});

// Handle login
document.getElementById("loginForm")?.addEventListener("submit", async function (e) {
    e.preventDefault();
    const formData = new FormData(this);
    const email = formData.get("email");
    const password = formData.get("password");

    try {
        const response = await fetch(`${API_BASE}/login`, {
            method: "POST",
            body: new URLSearchParams({ email, password }),
        });

        const result = await response.json();
        if (!response.ok) {
            alert(result.message || "Login failed");
            return;
        }

        localStorage.setItem("token", result.token); // Save JWT token
        alert("Login successful");
        window.location.href = "html/profile.html"; // Redirect to profile
    } catch (error) {
        console.error("Error during login:", error);
        alert("An error occurred. Please try again.");
    }
});
