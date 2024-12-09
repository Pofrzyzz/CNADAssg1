const API_USER = "/api/user";
const API_VEHICLE = "/api/vehicle";

let modifyReservationId = null; // Store the reservation ID being modified

// Fetch and display rental history
async function loadRentalHistory() {
    try {
        const token = localStorage.getItem("token");
        if (!token) {
            alert("You must be logged in to view your rental history.");
            window.location.href = "../index.html";
            return;
        }

        const response = await fetch(`http://localhost:8080/api/user/rental-history`, {
            method: "GET",
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });

        if (!response.ok) {
            alert("Failed to fetch rental history. Please try again later.");
            return;
        }

        const history = await response.json();
        console.log(history);
        if (history === null || history.length === 0) {
            alert("You have no rental history.");
            return;
        }

        const historyTable = document.getElementById("historyTable").querySelector("tbody");

        // Populate the rental history table
        historyTable.innerHTML = "";
        history.forEach((entry) => {
            const row = document.createElement("tr");

            row.innerHTML = `
                <td>${entry.vehicle_model}</td>
                <td>${new Date(entry.start_time).toLocaleString()}</td>
                <td>${new Date(entry.end_time).toLocaleString()}</td>
                <td>$${entry.total_cost.toFixed(2)}</td>
                <td>
                    <button class="btn modify-btn" data-id="${entry.reservation_id}">Modify</button>
                    <button 
                        class="btn cancel-btn" 
                        data-vehicle-id="${entry.vehicle_id}" 
                        data-start-time="${entry.start_time}" 
                        data-end-time="${entry.end_time}">
                        Cancel
                    </button>
                </td>
            `;

            historyTable.appendChild(row);
        });

        // Attach event listeners for Modify and Cancel buttons
        document.querySelectorAll(".modify-btn").forEach((button) => {
            button.addEventListener("click", () => openModifyModal(button.dataset.id));
        });

        document.querySelectorAll(".cancel-btn").forEach((button) => {
            button.addEventListener("click", () => {
                const vehicleModel = history[0].vehicle_model;
                console.log(vehicleModel);
                const startTime = button.dataset.startTime;
                console.log(startTime);
                const endTime = button.dataset.endTime;
                console.log(endTime); 
                cancelReservation(vehicleModel, startTime, endTime);
            });
        });
    } catch (error) {
        console.error("Error fetching rental history:", error);
        alert("An error occurred while loading your rental history.");
    }
}

// Open the modal for modifying a reservation
function openModifyModal(reservationId) {
    modifyReservationId = reservationId; // Set the reservation ID
    document.getElementById("modifyModal").style.display = "block";
    flatpickr(".datetime-picker", {
        enableTime: true,
        dateFormat: "Y-m-d H:i",
    });
}

// Close the modal
function closeModal() {
    document.getElementById("modifyModal").style.display = "none";
    modifyReservationId = null;
}

// Handle modify form submission
document.getElementById("modifyForm").addEventListener("submit", async function (e) {
    e.preventDefault();
    const newStartTime = document.getElementById("newStartTime").value;
    const newEndTime = document.getElementById("newEndTime").value;

    if (!newStartTime || !newEndTime) {
        alert("Please select both start and end times.");
        return;
    }

    try {
        const token = localStorage.getItem("token");
        if (!token) {
            alert("You must be logged in.");
            return;
        }

        const response = await fetch(`http://localhost:8080/api/vehicle/modify-booking`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                reservation_id: parseInt(modifyReservationId),
                start_time: toSQLFormat(newStartTime),
                end_time: toSQLFormat(newEndTime),
            }),
        });

        const result = await response.json();
        if (!response.ok) {
            alert(result.message || "Failed to modify reservation.");
            return;
        }

        alert("Reservation modified successfully!");
        closeModal();
        loadRentalHistory(); // Reload rental history
    } catch (error) {
        console.error("Error modifying reservation:", error);
        alert("An error occurred. Please try again.");
    }
});

// Cancel a reservation
async function cancelReservation(vehicleModel, startTime, endTime) {
    if (!confirm("Are you sure you want to cancel this reservation?")) {
        return;
    }

    try {
        const token = localStorage.getItem("token");
        if (!token) {
            alert("You must be logged in.");
            return;
        }

        // Fetch user ID
        const userIdResponse = await fetch("http://localhost:8080/api/user/user-id", {
            method: "GET",
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });

        const userIdData = await userIdResponse.json();
        if (!userIdResponse.ok) {
            alert("Failed to fetch user ID.");
            return;
        }


        // Retrieve the reservation ID from the booking result
        const vehIDResponse = await fetch(`http://localhost:8080/api/vehicle/retrieve-vehid`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                vehicle_model: vehicleModel,   
            }),
        });

        const vehIDData = await vehIDResponse.json();
        console.log(vehIDData);

        // Retrieve the reservation ID from the booking result
        const reservationResponse = await fetch(`http://localhost:8080/api/vehicle/find-reservationid`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                user_id: parseInt(userIdData.user_id),
                vehicle_id: parseInt(vehIDData.vehicle_id),
                start_time: toSQLFormat(startTime),
                end_time: toSQLFormat(endTime),     
            }),
        });

        const reservationData = await reservationResponse.json();
        if (!reservationResponse.ok) {
            alert(reservationData.message || "Failed to find reservation.");
            return;
        }

        const reservationId = reservationData.reservation_id;

        // Cancel the booking
        const cancelResponse = await fetch("http://localhost:8080/api/vehicle/cancel-booking", {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                reservation_id: parseInt(reservationId),
            }),
        });

        const cancelResult = await cancelResponse.json();
        if (!cancelResponse.ok) {
            alert(cancelResult.message || "Failed to cancel reservation.");
            return;
        }

        alert("Reservation canceled successfully.");
        loadRentalHistory(); // Reload rental history
    } catch (error) {
        console.error("Error canceling reservation:", error);
        alert("An error occurred. Please try again.");
    }
}

// Utility to convert ISO string to SQL format
function toSQLFormat(isoString) {
    return new Date(isoString).toISOString().slice(0, 19).replace("T", " ");
}

// Load rental history on page load
window.addEventListener("load", loadRentalHistory);
