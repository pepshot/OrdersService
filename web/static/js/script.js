function getOrder() {
    const orderId = document.getElementById('orderId').value.trim();
    if (!orderId) {
        showError('Please enter an Order ID');
        return;
    }

    // Clear previous results
    hideError();
    hideOrderInfo();
    showLoading();

    fetch(`/order/${orderId}`)
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => {
                    throw new Error(err.error || 'Failed to fetch order');
                });
            }
            return response.json();
        })
        .then(order => {
            hideLoading();
            displayOrder(order);
        })
        .catch(error => {
            hideLoading();
            showError(error.message);
        });
}

function displayOrder(order) {
    const orderDetails = document.getElementById('orderDetails');
    orderDetails.innerHTML = '';

    // Basic order info
    let html = `
        <div class="order-detail"><strong>Order UID:</strong> ${order.order_uid}</div>
        <div class="order-detail"><strong>Track Number:</strong> ${order.track_number}</div>
        <div class="order-detail"><strong>Customer ID:</strong> ${order.customer_id}</div>
        <div class="order-detail"><strong>Date Created:</strong> ${order.date_created}</div>
        <div class="order-detail"><strong>Delivery Service:</strong> ${order.delivery_service}</div>
    `;

    // Delivery info
    html += `
        <h3>Delivery Information</h3>
        <div class="order-detail"><strong>Name:</strong> ${order.delivery.name}</div>
        <div class="order-detail"><strong>Phone:</strong> ${order.delivery.phone}</div>
        <div class="order-detail"><strong>Address:</strong> ${order.delivery.address}, ${order.delivery.city}, ${order.delivery.region} ${order.delivery.zip}</div>
        <div class="order-detail"><strong>Email:</strong> ${order.delivery.email}</div>
    `;

    // Payment info
    html += `
        <h3>Payment Information</h3>
        <div class="order-detail"><strong>Transaction:</strong> ${order.payment.transaction}</div>
        <div class="order-detail"><strong>Amount:</strong> $${(order.payment.amount / 100).toFixed(2)}</div>
        <div class="order-detail"><strong>Currency:</strong> ${order.payment.currency}</div>
        <div class="order-detail"><strong>Provider:</strong> ${order.payment.provider}</div>
        <div class="order-detail"><strong>Payment Date:</strong> ${new Date(order.payment.payment_dt * 1000).toLocaleString()}</div>
    `;

    // Items
    if (order.items && order.items.length > 0) {
        html += `<h3>Items (${order.items.length})</h3><div class="items-list">`;
        order.items.forEach(item => {
            html += `
                <div class="item">
                    <h3>${item.name}</h3>
                    <div class="order-detail"><strong>Brand:</strong> ${item.brand}</div>
                    <div class="order-detail"><strong>Price:</strong> $${(item.price / 100).toFixed(2)}</div>
                    <div class="order-detail"><strong>Quantity:</strong> ${item.total_price / item.price}</div>
                    <div class="order-detail"><strong>Total:</strong> $${(item.total_price / 100).toFixed(2)}</div>
                    <div class="order-detail"><strong>Status:</strong> ${item.status}</div>
                </div>
            `;
        });
        html += `</div>`;
    }

    orderDetails.innerHTML = html;
    showOrderInfo();
}

function showLoading() {
    document.getElementById('loading').style.display = 'block';
}

function hideLoading() {
    document.getElementById('loading').style.display = 'none';
}

function showError(message) {
    const errorDiv = document.getElementById('error');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
}

function hideError() {
    document.getElementById('error').style.display = 'none';
}

function showOrderInfo() {
    document.getElementById('orderInfo').style.display = 'block';
}

function hideOrderInfo() {
    document.getElementById('orderInfo').style.display = 'none';
}

// Handle Enter key press
document.getElementById('orderId').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        getOrder();
    }
});