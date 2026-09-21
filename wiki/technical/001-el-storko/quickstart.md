# Quickstart: el-storko

## Backend

```bash
cd backend/el-storko
cp .env.example .env   # fill in JIRA_EMAIL, JIRA_API_TOKEN, JIRA_BASE_URL
go build -o bin/el-storko ./cmd/server
go run ./cmd/server -auto-migrate   # first run: creates schema + starts server on :8082
```

Migration control (matches `oncarinho`/`hustle-turtle`):

```bash
go run ./cmd/server -migrate=up
go run ./cmd/server -migrate=down
go run ./cmd/server -version
```

## CLI

```bash
cd backend/el-storko
go build -o bin/el-storko-cli ./cmd/cli
./bin/el-storko-cli add "Buy birthday gift"
./bin/el-storko-cli list --status todo
```

## Frontend

```bash
pnpm --filter el-storko dev   # http://localhost:3101
```

## Always-on (launchd)

```bash
cp infra/launchd/com.movoz.el-storko-backend.plist ~/Library/LaunchAgents/
cp infra/launchd/com.movoz.el-storko-frontend.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.movoz.el-storko-backend.plist
launchctl load ~/Library/LaunchAgents/com.movoz.el-storko-frontend.plist
```

Verify survival: `kill -9 <pid>` the backend process, confirm launchd restarts it within a few
seconds; reboot the machine, confirm both services are reachable without manual intervention.

## Smoke test

```bash
curl -X POST localhost:8082/api/work-items -d '{"type":"epic","title":"Home projects"}' -H 'Content-Type: application/json'
curl localhost:8082/api/work-items
```
