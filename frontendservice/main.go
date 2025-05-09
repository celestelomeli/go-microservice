package main

import (
	"fmt"      // For printing formatted strings to output
	"log"      // For logging info and errors 
	"net/http" // For running an HTTP server in Go
)

func main() {
	// Register handler function for the root URL "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set content type of the HTTP response to HTML
		w.Header().Set("Content-Type", "text/html")

		// Send HTML + CSS + JavaScript as the HTTP response
		fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Go App Frontend</title>
	<style>
		/* General styling */
		body {
			font-family: Arial;
			padding: 2rem;
			background: #f9f9f9;
		}

		/* Form styling */
		form {
			background: white;
			padding: 1rem;
			margin-bottom: 2rem;
			border-radius: 8px;
			box-shadow: 0 2px 5px rgba(0,0,0,0.1);
			max-width: 400px;
		}

		/* Label spacing */
		label {
			display: block;
			margin-top: 1rem;
		}

		/* Input and dropdown styling */
		input, select {
			width: 100%;
			padding: 0.5rem;
			margin-top: 0.3rem;
		}

		/* Button styling */
		button {
			margin-top: 1rem;
			padding: 0.7rem;
			background-color: #007bff;
			color: white;
			border: none;
			border-radius: 4px;
			cursor: pointer;
		}

		/* <pre> tag formats long text, keeps spacing */
		pre {
			background: #eee;
			padding: 1rem;
			border-radius: 6px;
			white-space: pre-wrap; /* Preserve whitespace and wrap long lines */
		}
	</style>
</head>
<body>

	<!-- USER CREATION FORM -->
	<h2>Create a User</h2>
	<form id="userForm">
		<label>Name: <input type="text" id="name" required /></label>
		<label>Email: <input type="email" id="email" required /></label>
		<button type="submit">Create User</button>
	</form>
	<pre id="userResponse"></pre> <!-- Shows JSON response of created user -->

	<!-- ORDER CREATION FORM -->
	<h2>Place an Order</h2>
	<form id="orderForm">
		<label>User:
			<select id="user_id"></select> <!-- Populated with user list -->
		</label>
		<label>Product:
			<select id="product_id"></select> <!-- Populated with product list -->
		</label>
		<label>Quantity:
			<input type="number" id="quantity" value="1" min="1" required />
		</label>
		<button type="submit">Submit Order</button>
	</form>
	<pre id="orderResponse"></pre> <!-- JSON response of created order -->

	<!-- LOAD ALL ORDERS -->
	<h2>View All Orders</h2>
	<button id="loadOrdersBtn">Load Orders</button>
	<pre id="ordersList"></pre> <!-- Shows full list of orders -->

<script>
	// Load all users from backend and populate user dropdown
	async function loadUsers() {
		const res = await fetch("http://localhost:8080/users");
		const users = await res.json();
		const userSelect = document.getElementById("user_id");
		userSelect.innerHTML = ""; // Clears existing options

		// Add each user as a <option>
		users.forEach(u => {
			const option = document.createElement("option");
			option.value = u.id;
			option.textContent = u.name + " (" + u.email + ")";
			userSelect.appendChild(option);
		});
	}

	// Load all products from backend and populate product dropdown
	async function loadProducts() {
		const res = await fetch("http://localhost:8080/products");
		const products = await res.json();
		const select = document.getElementById("product_id");
		select.innerHTML = ""; // Clears existing options

		// Add each product as a <option>
		products.forEach(p => {
			const option = document.createElement("option");
			option.value = p.id;
			option.textContent = p.name + " ($" + p.price + ")";
			select.appendChild(option);
		});
	}

	// On user form submit, create new user
	document.getElementById("userForm").addEventListener("submit", async function(e) {
		e.preventDefault(); // Prevent default form submission

		const name = document.getElementById("name").value;
		const email = document.getElementById("email").value;

		const res = await fetch("http://localhost:8080/users", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ name, email }) // Send data as JSON
		});

		const data = await res.json();

		// Show formatted JSON response in <pre> tag
		document.getElementById("userResponse").textContent = JSON.stringify(data, null, 2);

		await loadUsers(); // Reload users to update dropdown
	});

	// On order form submit, create new order
	document.getElementById("orderForm").addEventListener("submit", async function(e) {
		e.preventDefault();

		const user_id = Number(document.getElementById("user_id").value);
		const product_id = Number(document.getElementById("product_id").value);
		const quantity = Number(document.getElementById("quantity").value);

		const res = await fetch("http://localhost:8080/orders", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ user_id, product_id, quantity })
		});

		const data = await res.json();

		// Pretty print JSON (null = include all fields, 2 = indent spacing)
		document.getElementById("orderResponse").textContent = JSON.stringify(data, null, 2);
	});

	// On click "Load Orders" button, fetch all orders
	document.getElementById("loadOrdersBtn").addEventListener("click", async () => {
		const res = await fetch("http://localhost:8080/orders");
		const data = await res.json();
		document.getElementById("ordersList").textContent = JSON.stringify(data, null, 2);
	});

	// Load users and products when the page first loads
	loadUsers();
	loadProducts();
</script>
</body>
</html>
`)
	})

	// Log to console and run server on port 3000
	log.Println("Frontend service running on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil)) // Crash on error
}
