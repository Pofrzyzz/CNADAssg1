const API_BASE = "/api/user"; // Adjust this if needed

// Fetch and display user profile details
async function loadProfile() {
    try {
        const token = localStorage.getItem("token");
        if (!token) {
            alert("You must be logged in to view your profile.");
            window.location.href = "../index.html";
            return;
        }

        const response = await fetch(`http://localhost:8080/api/user/profile`, {
            method: "GET",
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });

        if (!response.ok) {
            alert("Failed to fetch profile. Please log in again.");
            localStorage.removeItem("token");
            window.location.href = "../index.html";
            return;
        }

        const profile = await response.json();
        document.getElementById("profileEmail").textContent = profile.email;
        document.getElementById("profilePhone").textContent = profile.phone_number;
        document.getElementById("profileMembership").textContent = profile.membership_tier;
    } catch (error) {
        console.error("Error fetching profile:", error);
        alert("An error occurred while loading your profile.");
    }
}

// Edit and update user details (only phone number)
document.getElementById("editDetailsBtn")?.addEventListener("click", async function () {
    const newPhone = prompt("Enter your new phone number:");
    if (!newPhone) {
        alert("Phone number cannot be empty.");
        return;
    }

    try {
        const token = localStorage.getItem("token");
        if (!token) {
            alert("You must be logged in to update your details.");
            window.location.href = "../index.html";
            return;
        }

        // Send a request to update the phone number
        const response = await fetch(`http://localhost:8080/api/user/update-profile`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded",
                Authorization: `Bearer ${token}`,
            },
            body: new URLSearchParams({ phone_number: newPhone }),
        });

        const result = await response.json();
        if (!response.ok) {
            alert(result.message || "Failed to update details.");
            return;
        }

        alert("Phone number updated successfully!");

        // Now request OTP for phone number verification
        const otpResponse = await fetch(`http://localhost:8080/api/user/generate-otp`, {
            method: "POST",
            body: new URLSearchParams({ phone_number: newPhone }),
        });

        const otpResult = await otpResponse.json();
        if (!otpResponse.ok) {
            alert(otpResult.message || "Failed to generate OTP");
            return;
        }

        // Prompt the user for the OTP
        const userOTP = prompt("Enter the OTP sent to your new phone number:");
        if (!userOTP) {
            alert("OTP is required to proceed.");
            return;
        }

        // Verify OTP
        const verifyResponse = await fetch(`http://localhost:8080/api/user/verify-otp`, {
            method: "POST",
            body: new URLSearchParams({ phone_number: newPhone, otp: userOTP }),
        });

        const verifyResult = await verifyResponse.json();
        if (!verifyResponse.ok) {
            alert(verifyResult.message || "OTP verification failed");
            return;
        }

        alert("Phone number successfully verified!");
        loadProfile(); // Reload the profile with updated details
    } catch (error) {
        console.error("Error updating profile:", error);
        alert("An error occurred while updating your details.");
    }
});

// Navigation buttons
document.getElementById("viewInvoicesBtn")?.addEventListener("click", function () {
    window.location.href = "invoices.html";
});

document.getElementById("viewHistoryBtn")?.addEventListener("click", function () {
    window.location.href = "history.html";
});

// Logout button
document.querySelector(".logoutBtn")?.addEventListener("click", function () {
    localStorage.removeItem("token");
    window.location.href = "../index.html";
});

// Load profile details on page load
window.addEventListener("load", loadProfile);
