const API_VEHICLE = "/api/vehicle";
const API_BILLING = "/api/billing";

// Load available vehicles and populate the booking page
async function loadVehicles() {
    try {
        const response = await fetch(`http://localhost:8080/api/vehicle/vehicles?availability=true`, {
            method: "GET",
        });

        if (!response.ok) {
            alert("Failed to load available vehicles.");
            return;
        }

        const vehicles = await response.json();
        const vehicleSelect = document.getElementById("vehicleSelect");
        const vehicleList = document.getElementById("vehicleList");
        if (vehicles === null || vehicles.length === 0) {
            vehicleSelect.innerHTML = "<option value=''>No vehicles available</option>";
            vehicleList.innerHTML = "<p>No vehicles available</p>";
            return;
        }

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
    console.log(formData);
    const vehicleID = formData.get("vehicle_id");
    console.log(vehicleID);
    const startTime = formData.get("start_time");
    console.log(startTime);
    const endTime = formData.get("end_time");
    console.log(endTime);

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

        // Fetch user ID using the /user-id endpoint
        const userIdResponse = await fetch("http://localhost:8080/api/user/user-id", {
            method: "GET",
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });

        const userIdData = await userIdResponse.json();
        console.log('User ID:', userIdData);
        console.log(typeof(userIdData.user_id));
        console.log(typeof(vehicleID));
        console.log(typeof(startTime));
        console.log(typeof(endTime));

        // Estimate cost before booking
        const estimateResponse = await fetch(`http://localhost:8080/api/billing/calculate-cost`, {
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
        const Amount = estimate.estimated_cost.toFixed(2);
        if (!confirmBooking) return;

        // Make the booking
        const bookingResponse = await fetch(`http://localhost:8080/api/vehicle/book-vehicle`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                user_id: parseInt(userIdData.user_id),
                vehicle_id: parseInt(vehicleID),
                start_time: toSQLFormat(startTime),
                end_time: toSQLFormat(endTime),
            }),
        });

        const bookingResult = await bookingResponse.json();
        if (!bookingResponse.ok) {
            alert(bookingResult.message || "Failed to book the vehicle.");
            return;
        }

        // Retrieve the reservation ID from the booking result
        const reservationResponse = await fetch(`http://localhost:8080/api/vehicle/find-reservationid`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                user_id: parseInt(userIdData.user_id),
                vehicle_id: parseInt(vehicleID),
                start_time: toSQLFormat(startTime),
                end_time: toSQLFormat(endTime),     
            }),
        });

        const reservationData = await reservationResponse.json();
        console.log(reservationData);
        console.log('Reservation Data:', reservationData.reservation_id);
        PaymentStatus = "Paid"; // Hardcoded payment status for now as theres no official payment method
        InvoiceID = generateInvoiceID(reservationData.reservation_id);
        console.log('Invoice ID:', InvoiceID);

        const Amount2 = parseFloat(Amount);
        // Make the payment
        const paymentResponse = await fetch(`http://localhost:8080/api/billing/make-payment`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                reservation_id: parseInt(reservationData.reservation_id),
                amount: parseFloat(Amount),
                payment_status: PaymentStatus,
                invoice_id: InvoiceID,
            }),
        });

        const paymentResult = await paymentResponse.json();
        if (!paymentResponse.ok) {
            alert(paymentResult.message || "Failed to make payment.");
            return;
        }
        console.log(parseInt(vehicleID));
        // Retrieve the vehicle model based on the vehicle ID
        const modelResponse = await fetch(`http://localhost:8080/api/vehicle/retrieve-model`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                vehicle_id: parseInt(vehicleID),
            }),
        });
        console.log(modelResponse);
        const modelData = await modelResponse.json();
        console.log(modelData);
        if (!modelResponse.ok) {
            alert(modelData.message || "Failed to retrieve vehicle model.");
            return;
        }
        console.log(modelData);
        console.log("Retrieved Vehicle Model:", modelData.vehicle_model);
        console.log(typeof(modelData.vehicle_model));
        console.log({
            user_id: parseInt(userIdData.user_id),
            vehicle_model: modelData.vehicle_model, // Assuming this is fetched correctly
            start_time: toSQLFormat(startTime),
            end_time: toSQLFormat(endTime),
            total_cost: parseFloat(Amount),
        });
        console.log(vehicleID);
        console.log(typeof(vehicleID));
        // Update rental history
        const historyResponse = await fetch(`http://localhost:8080/api/vehicle/update-history`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                user_id: parseInt(userIdData.user_id),
                vehicle_id: parseInt(vehicleID),
                start_time: toSQLFormat(startTime),
                end_time: toSQLFormat(endTime),
                total_cost: Amount2,
            }),
        });
        console.log(historyResponse);
        const historyResult = await historyResponse.json();
        if (!historyResponse.ok) {
            alert(historyResult.message || "Failed to update rental history.");
            return;
        }

        console.log("Rental history updated:", historyResult.message);


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

// utility to convert ISO string to SQL format
function toSQLFormat(isoString) {
    return new Date(isoString).toISOString().slice(0, 19).replace("T", " ");
}

function generateInvoiceID(reservationID) {
    // Get the current date
    const now = new Date();

    // Format the date as YYYYMMDD
    const year = now.getFullYear(); // Get the year (YYYY)
    const month = String(now.getMonth() + 1).padStart(2, '0'); // Get the month (MM) and pad to 2 digits
    const day = String(now.getDate()).padStart(2, '0'); // Get the day (DD) and pad to 2 digits

    // Construct the invoice ID
    const invoiceID = `INV-${year}${month}${day}-${reservationID}`;

    return invoiceID;
}


// Load vehicles on page load
window.addEventListener("load", loadVehicles);
