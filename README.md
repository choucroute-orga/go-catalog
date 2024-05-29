# Go Catalog

This MS stores the ingredients.

### Start the server

```bash
cp .env.example .env
export $(cat .env | xargs)
docker-compose up
go run main.go
```