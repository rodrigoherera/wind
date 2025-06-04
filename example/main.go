package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Simple HTTP handlers
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/time", timeHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8080"
	fmt.Printf("🌪️ Wind Example Server starting on http://localhost%s\n", port)
	fmt.Println("Try editing this file and watch Wind reload automatically!")

	log.Fatal(http.ListenAndServe(port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Wind Example</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .emoji { font-size: 2em; }
        .button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; margin: 5px; }
        .button:hover { background: #0056b3; }
        #time { font-family: monospace; font-size: 1.2em; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1><span class="emoji">🌪️</span> Wind Example Server</h1>
        <p>This is a simple Go web application running with Wind!</p>
        <p>Try editing <code>main.go</code> and watch the server reload automatically.</p>
        
        <h3>Current Server Time:</h3>
        <div id="time">Loading...</div>
        
        <br>
        <button class="button" onclick="loadTime()">Refresh Time</button>
        <button class="button" onclick="window.location.href='/health'">Health Check</button>
        
        <h3>Features:</h3>
        <ul>
            <li>🔄 Auto-reload on file changes</li>
            <li>⚡ Fast rebuilds</li>
            <li>🎨 Colored terminal output</li>
            <li>🚀 Zero configuration</li>
        </ul>
    </div>

    <script>
        function loadTime() {
            fetch('/api/time')
                .then(response => response.text())
                .then(time => {
                    document.getElementById('time').textContent = time;
                })
                .catch(error => {
                    document.getElementById('time').textContent = 'Error loading time';
                });
        }
        
        // Load time on page load
        loadTime();
        
        // Auto-refresh time every 5 seconds
        setInterval(loadTime, 5000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Server time: %s", time.Now().Format("2006-01-02 15:04:05"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status": "healthy", "service": "wind-example", "timestamp": "`+time.Now().Format(time.RFC3339)+`"}`)
}
