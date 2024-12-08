document.getElementById('loginForm')?.addEventListener('submit', async function (e) {
    e.preventDefault();
    const formData = new FormData(this);
    const response = await fetch('/user-service/login', { method: 'POST', body: formData });
    const result = await response.json();
    if (result.success) window.location.href = '/dashboard.html';
    else alert(result.message);
});

document.getElementById('registerForm')?.addEventListener('submit', async function (e) {
    e.preventDefault();
    const formData = new FormData(this);
    const response = await fetch('/register', { method: 'POST', body: formData });
    const result = await response.json();
    alert(result.message);
});

async function viewVehicles() {
    const response = await fetch('/vehicle-service/vehicles');
    const vehicles = await response.json();
    document.getElementById('vehicleList').innerHTML = vehicles.map(
        v => `<div>${v.model} - ${v.license_plate} (${v.availability ? 'Available' : 'Unavailable'})</div>`
    ).join('');
}

async function viewInvoices() {
    const response = await fetch('/billing-service/invoices/user/1'); // Replace '1' with logged-in user_id
    const invoices = await response.json();
    document.getElementById('invoiceList').innerHTML = invoices.map(
        i => `<div>Reservation: ${i.reservation_id}, Amount: $${i.amount}, Status: ${i.payment_status}</div>`
    ).join('');
}

function logout() {
    alert('Logged out');
    window.location.href = '/';
}
