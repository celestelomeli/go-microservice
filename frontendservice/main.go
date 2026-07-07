package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Go App Frontend</title>
	<style>
		body {
			font-family: Arial;
			padding: 2rem;
			background: #f9f9f9;
		}

		form {
			background: white;
			padding: 1rem;
			margin-bottom: 2rem;
			border-radius: 8px;
			box-shadow: 0 2px 5px rgba(0,0,0,0.1);
			max-width: 400px;
		}

		label {
			display: block;
			margin-top: 1rem;
		}

		input, select {
			width: 100%;
			padding: 0.5rem;
			margin-top: 0.3rem;
		}

		button {
			margin-top: 1rem;
			padding: 0.7rem;
			background-color: #007bff;
			color: white;
			border: none;
			border-radius: 4px;
			cursor: pointer;
		}

		pre {
			background: #eee;
			padding: 1rem;
			border-radius: 6px;
			white-space: pre-wrap;
		}
	</style>
</head>
<body>

	<h2>Create a User</h2>
	<form id="userForm">
		<label>Name: <input type="text" id="name" required /></label>
		<label>Email: <input type="email" id="email" required /></label>
		<button type="submit">Create User</button>
	</form>
	<pre id="userResponse"></pre>

	<h2>Place an Order</h2>
	<form id="orderForm">
		<label>User:
			<select id="user_id"></select>
		</label>
		<label>Product:
			<select id="product_id"></select>
		</label>
		<label>Quantity:
			<input type="number" id="quantity" value="1" min="1" required />
		</label>
		<button type="submit">Submit Order</button>
	</form>
	<pre id="orderResponse"></pre>

	<h2>View All Orders</h2>
	<button id="loadOrdersBtn">Load Orders</button>
	<pre id="ordersList"></pre>

<script>
	async function loadUsers() {
		const res = await fetch("http://localhost:8080/users");
		const users = await res.json();
		const userSelect = document.getElementById("user_id");
		userSelect.innerHTML = "";

		users.forEach(u => {
			const option = document.createElement("option");
			option.value = u.id;
			option.textContent = u.name + " (" + u.email + ")";
			userSelect.appendChild(option);
		});
	}

	async function loadProducts() {
		const res = await fetch("http://localhost:8080/products");
		const products = await res.json();
		const select = document.getElementById("product_id");
		select.innerHTML = "";

		products.forEach(p => {
			const option = document.createElement("option");
			option.value = p.id;
			option.textContent = p.name + " ($" + p.price + ")";
			select.appendChild(option);
		});
	}

	document.getElementById("userForm").addEventListener("submit", async function(e) {
		e.preventDefault();

		const name = document.getElementById("name").value;
		const email = document.getElementById("email").value;

		const res = await fetch("http://localhost:8080/users", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ name, email })
		});

		const data = await res.json();
		document.getElementById("userResponse").textContent = JSON.stringify(data, null, 2);

		await loadUsers();
	});

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
		document.getElementById("orderResponse").textContent = JSON.stringify(data, null, 2);
	});

	document.getElementById("loadOrdersBtn").addEventListener("click", async () => {
		const res = await fetch("http://localhost:8080/orders");
		const data = await res.json();
		document.getElementById("ordersList").textContent = JSON.stringify(data, null, 2);
	});

	loadUsers();
	loadProducts();
</script>
</body>
</html>
`)
	})

	log.Println("Frontend service running on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
