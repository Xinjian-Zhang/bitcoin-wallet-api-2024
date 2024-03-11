# Bitcoin Wallet API

## Overview

The Bitcoin Wallet API is a simple solution for managing Bitcoin transactions within a wallet application. Built with Go, it utilizes the Gin web framework for handling HTTP requests and GORM for database interactions, specifically with SQLite for storage. 

## Getting Started

### Prerequisites

The application is built and tested with Go version 1.22.1, its dependency on Go Modules for managing external libraries.

### Setup

1. **Clone the Repo**

   ```bash
   git clone https://github.com/Xinjian-Zhang/bitcoin-wallet-api-2024.git
   cd bitcoin-wallet-api-2024/bitcoin-wallet-api
   ```

2. **Install Dependencies**:
   The application requires several Go packages, including Gin for web routing and GORM for database operations. 

   To install these dependencies, run:

   ```bash
   go mod tidy
   ```
   

### Running the Application

Launch the server by executing:

```bash
go run main/main.go
```

The API available at `http://localhost:8080`.

### Available Endpoints

- **GET `/transactions`**: Lists all transactions that are unspent.
- **GET `/balance`**: Displays the wallet's balance in BTC and the equivalent in EUR.
- **POST `/transfer`**: Initiates a new transfer. Requires a JSON payload with an `amount_eur` field specifying the transfer amount in Euros.

---

### Contact

**Xinjian Zhang**

[1xinjian.zhang@gmail.com](mailto:1xinjian.zhang@gmail.com)

