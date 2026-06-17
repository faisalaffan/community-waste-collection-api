# Contributing

## Setup

```bash
git clone https://github.com/faisalaffan/community-waste-collection-api.git
cd community-waste-collection-api
cp .env.example .env
make docker-up
make schema-apply
```

## Development workflow

1. Create a feature branch from `dev`
2. Write tests for your changes
3. Implement your changes
4. Run `make lint test` to verify
5. Open a pull request to `dev`

## Code conventions

- Go standard formatting (`gofmt`)
- Follow existing layering: repository → service → handler
- Interfaces for testability
- No GORM AutoMigrate — use Atlas for schema changes
