
## Installation

```bash

# 1. Clean existing dependencies
rm go.sum

# 2. Reinstall dependencies
go mod tidy
go mod download

# 3. Configure environment variables
nano .env

# 4. Run the application
go run main.go

```


