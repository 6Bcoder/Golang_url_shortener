# URL Shortener

A simple URL shortener application built in Go using SQLite as the database for storing long and short URLs.

## Features

- Shorten long URLs into short, unique URLs.
- Redirect to the original long URL when a short URL is visited.
- Track and display the top domains based on visits.

## Requirements

- Go (1.x+)
- SQLite (for storing URL mappings)
## Setup

### Clone the Repository

```bash
git clone https://github.com/6Bcoder/Golang_url_short.git
cd Golang_url_short

1. Install the required Go dependencies:

2. Set Up SQLite Database:
Make sure you have SQLite installed on your machine. The database for this project is stored in urlshortener.db, which you can change by editing the dbFile constant in the main.go file.
Ensure the urlshortener.db file exists, or the project will create it automatically upon running.

3. Run the Project:
To run the server, use the following command:
go run main.go
The server will start on http://localhost:3000.

4. Interact with the API:
You can interact with the URL Shortener API using the following endpoints:
POST /shorten: Shorten a URL.
Request Body (JSON):
{
  "longURL": "https://example.com"
}
Response Body (JSON):
{
  "shortURL": "swew5rgg"
}
GET /redirect/{shortURL}: Redirect to the original long URL.
Example: http://localhost:3000/redirect/{shortURL}.
GET /metrics: Retrieve metrics for the most visited domains.
