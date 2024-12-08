const API_USER = "/api/user";

// Fetch and display rental history
async function loadRentalHistory() {
    try {
        const token = localStorage.getItem("token");
        if (!token) {
            alert("You must be logged in to view your rental history.");
            window.location.href = "../index.html";
            return;
        }

        const response = await fetch(`${API_USER}/rental-history`, {
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
            `;

            historyTable.appendChild(row);
        });
    } catch (error) {
        console.error("Error fetching rental history:", error);
        alert("An error occurred while loading your rental history.");
    }
}

// Load rental history on page load
window.addEventListener("load", loadRentalHistory);
