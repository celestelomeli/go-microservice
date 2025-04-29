package main

import (
	"fmt"          // For formatted I/O    
	"log"          // For logging server events and errors
	"net/http"     // For building HTTP web server
)

// Register a route handler function for root path "/"
// Whenever someone accesses the root, function will be triggered 
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set response header to tell browser sending back HTML
		w.Header().Set("Content-Type", "text/html")
		
		// Write HTML content into the HTTP response/browser
		// This will render the webpage
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Welcome</title>
				<style>
					/* Style the whole page */
					body {
						display: flex;                        /* Use Flexbox to center content */
						justify-content: center;              /* Center horizontally */
						align-items: center;                  /* Center vertically */
						height: 100vh;                        /* Full height */
						margin: 0;                            /* Remove default margin */
						font-family: Arial, sans-serif;       /* Set font */
						background: linear-gradient(-45deg, #2193b0, #6dd5ed, #b2fefa, #0f2027); */ Gradient background */
						background-size: 400% 400%;           /* Make background larger for animation */
						animation: gradientBG 15s ease infinite;  /* Animate background */
					}
					/* Style the heading (h1) */	
					h1 {
						color: #fff;                                /* White text */
						font-size: 3em;
						text-shadow: 4px 4px 8px rgba(0, 0, 0, 0.6); /* Soft shadow behind text for better readability */
						opacity: 0;                                  /* Start invisible */
						animation: fadeIn 2s ease-in-out forwards;   /* Fade in animation over 2 seconds */
					}
					/* Define the fade-in animation for the h1 */
					@keyframes fadeIn {
						from { opacity: 0; }    /* Start fully transparent */
						to { opacity: 1; }      /* End fully opaque */
					}
					/* Define the background animation for the body */	
					@keyframes gradientBG {
						0% { background-position: 0% 50%; }    /* Start on left */
						50% { background-position: 100% 50%; } /* Move to right */
						100% { background-position: 0% 50%; }  /* Move back to left */
					}
				</style>
			</head>
			<body>
				<h1>Welcome to my Go App</h1> <!-- This is the big welcome text -->
			</body>
			</html>
		`)
	})
	// Log that the server has started
	log.Println("Frontend service running on :3000")
	// Start the HTTP server on port 3000
	// If there is an error starting the server, log the error and exit
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
