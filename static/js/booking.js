const API_VEHICLE = "/api/vehicle";
const API_BILLING = "/api/billing";

// Load available vehicles and populate the booking page
async function loadVehicles() {
    try {
        const response = await fetch(`${API_VEHICLE}/vehicles?availability=true`, {
            method: "GET",
        });

        if (!response.ok) {
            alert("Failed to load available vehicles.");
            return;
        }

        const vehicles = await response.json();
        const vehicleSelect = document.getElementById("vehicleSelect");
        const vehicleList = document.getElementById("vehicleList");

        // Populate the dropdown
        vehicles.forEach((vehicle) => {
            const option = document.createElement("option");
            option.value = vehicle.vehicle_id;
            option.textContent = `${vehicle.model} (${vehicle.license_plate}) - ${vehicle.location}`;
            vehicleSelect.appendChild(option);
        });

        // Populate the vehicle list
        vehicleList.innerHTML = vehicles
            .map(
                (vehicle) =>
                    `<p><strong>${vehicle.model}</strong> - License: ${vehicle.license_plate}, Location: ${vehicle.location}, Charge Level: ${vehicle.charge_level}%</p>`
            )
            .join("");
    } catch (error) {
        console.error("Error loading vehicles:", error);
        alert("An error occurred while loading vehicles.");
    }
}

// Handle booking submission
document.getElementById("bookingForm")?.addEventListener("submit", async function (e) {
    e.preventDefault();

    const formData = new FormData(this);
    const vehicleID = formData.get("vehicle_id");
    const startTime = formData.get("start_time");
    const endTime = formData.get("end_time");

    if (!vehicleID || !startTime || !endTime) {
        alert("All fields are required.");
        return;
    }

    try {
        const token = localStorage.getItem("token");
        if (!token) {
            alert("You must be logged in to make a booking.");
            window.location.href = "../index.html";
            return;
        }

        // Estimate cost before booking
        const estimateResponse = await fetch(`${API_BILLING}/calculate-cost`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                membership_tier: "Basic", // Replace with actual user tier if available
                hours: calculateHours(startTime, endTime),
                discount: 10, // Example discount, replace as needed
            }),
        });

        const estimate = await estimateResponse.json();
        if (!estimateResponse.ok) {
            alert(estimate.message || "Failed to calculate estimated cost.");
            return;
        }

        const confirmBooking = confirm(
            `Estimated Cost: $${estimate.estimated_cost.toFixed(2)}. Do you want to proceed?`
        );
        if (!confirmBooking) return;

        // Make the booking
        const bookingResponse = await fetch(`${API_VEHICLE}/book-vehicle`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                vehicle_id: vehicleID,
                start_time: startTime,
                end_time: endTime,
            }),
        });

        const bookingResult = await bookingResponse.json();
        if (!bookingResponse.ok) {
            alert(bookingResult.message || "Failed to book the vehicle.");
            return;
        }

        alert("Booking successful!");
        window.location.href = "profile.html"; // Redirect to profile after booking
    } catch (error) {
        console.error("Error during booking:", error);
        alert("An error occurred. Please try again.");
    }
});

// Utility to calculate hours between two timestamps
function calculateHours(start, end) {
    const startTime = new Date(start);
    const endTime = new Date(end);
    const diffInMilliseconds = endTime - startTime;
    return diffInMilliseconds / (1000 * 60 * 60); // Convert milliseconds to hours
}

// Load vehicles on page load
window.addEventListener("load", loadVehicles);
