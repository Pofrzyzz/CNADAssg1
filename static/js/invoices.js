const API_BILLING = "/api/billing";

// Fetch and display user invoices
async function loadInvoices() {
    try {
        const token = localStorage.getItem("token");
        console.log(token);

        if (!token) {
            alert("You must be logged in to view invoices.");
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
        console.log(userIdResponse);
        if (!userIdResponse.ok) {
            alert("Failed to retrieve user information. Please log in again.");
            localStorage.removeItem("token");
            window.location.href = "../index.html";
            return;
        }

        const userIdData = await userIdResponse.json();
        const userID = userIdData.user_id; // Extract the user_id from the response

        // Fetch invoices for the user using the user ID
        const response = await fetch(`http://localhost:8080/api/billing/invoices/user/${userID}`, {
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
