const API_BILLING = "/api/billing";

// Fetch and display user invoices
async function loadInvoices() {
    try {
        const token = localStorage.getItem("token");
        if (!token) {
            alert("You must be logged in to view invoices.");
            window.location.href = "../index.html";
            return;
        }

        const userID = localStorage.getItem("user_id"); // Assume `user_id` is stored during login
        if (!userID) {
            alert("Failed to retrieve user information. Please log in again.");
            localStorage.removeItem("token");
            window.location.href = "../index.html";
            return;
        }

        const response = await fetch(`${API_BILLING}/invoices/user/${userID}`, {
            method: "GET",
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });

        if (!response.ok) {
            alert("Failed to fetch invoices. Please try again later.");
            return;
        }

        const invoices = await response.json();
        const invoiceTable = document.getElementById("invoiceTable").querySelector("tbody");

        // Populate the invoice table
        invoiceTable.innerHTML = "";
        invoices.forEach((invoice) => {
            const row = document.createElement("tr");

            row.innerHTML = `
                <td>${invoice.reservation_id}</td>
                <td>$${invoice.amount.toFixed(2)}</td>
                <td>${invoice.payment_status}</td>
                <td>${new Date(invoice.created_at).toLocaleString()}</td>
            `;

            invoiceTable.appendChild(row);
        });
    } catch (error) {
        console.error("Error fetching invoices:", error);
        alert("An error occurred while loading invoices.");
    }
}

// Load invoices on page load
window.addEventListener("load", loadInvoices);
