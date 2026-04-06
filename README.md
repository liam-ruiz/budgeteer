# Budgeting

A local-first budgeting app with a Go backend, Angular frontend, Postgres database, and Plaid account linking. App data is stored in your local Postgres Docker volume unless you reset or remove it.

## Quick Start

### 1. Install and open Docker

Install Docker Desktop from the official Docker docs:

- Mac, Windows, and Linux: https://docs.docker.com/get-started/introduction/get-docker-desktop/

After installing it, open Docker Desktop and wait until it says Docker is running. If Docker is not open, `docker compose` commands can fail with a Docker daemon connection error.

You can check from a terminal:

```sh
docker info
docker compose version
```

### 2. Create your local environment file

Copy the example file and fill in your secrets:

```sh
cp .env.example .env
```

At minimum, confirm these values in `.env`:

```sh
DATABASE_URL=postgres://postgres:mysecretpassword@db:5432/budget
JWT_SECRET=replace-with-a-long-random-secret-at-least-32-characters
PLAID_CLIENT_ID=<YOUR-CLIENT-ID>
PLAID_SECRET=<YOUR-SECRET>
PLAID_ENV=sandbox
BACKEND_PORT=8080
RESET_DB=false
ALLOWED_ORIGINS=http://localhost:3000
BASE_URL=http://localhost
WEBHOOK_URL=<YOUR-TUNNEL-URL>

API_URL=http://localhost:8080
FRONTEND_PORT=3000

POSTGRES_USER=postgres
POSTGRES_PASSWORD=mysecretpassword
POSTGRES_DB=budget
DB_PORT=5432
```

Notes:

- `DATABASE_URL` uses `db` as the hostname because Docker Compose runs Postgres as the `db` service.
- `JWT_SECRET` should be a long random value. Do not reuse the example value outside local development.
- `RESET_DB=true` rolls back and reapplies migrations when the backend starts, which deletes existing local data. For a brand-new local database, you may need to run once with `RESET_DB=true`, then switch it back to `false` before continuing.
- `ALLOWED_ORIGINS` must include the frontend origin. With the default Docker setup, use `http://localhost:3000`.
- `WEBHOOK_URL` should be only the public origin, such as `https://abc123.ngrok-free.app`. Do not add `/api/plaid/webhook`; the backend appends that path when creating a Plaid Link token.
- `API_URL` and `BASE_URL` are included in the env example for local configuration, but the current Docker frontend build uses `frontend/src/environments/environment.ts`, which points to `http://localhost:8080/api`. If you change `BACKEND_PORT`, update that frontend environment file or build configuration too.

### 3. Get Plaid credentials

Create or log in to your Plaid developer account:

- Plaid Dashboard: https://dashboard.plaid.com/
- Plaid Quickstart docs: https://plaid.com/docs/quickstart/
- Plaid Sandbox docs: https://plaid.com/docs/sandbox/

In the Plaid Dashboard, open the API keys area and copy:

- `client_id`: the private identifier for your Plaid team.
- `secret`: the private key for the environment you are using.

For local development, start with:

```sh
PLAID_CLIENT_ID=<your-apps-client-id>
PLAID_ENV=sandbox
PLAID_SECRET=<your-sandbox-secret>
```

Make sure that this is done in `.env` and not `.env.example`. Never commit a real Plaid secret to git. 

Sandbox vs. production:

- `sandbox` uses Plaid test institutions and test data. It is the right default for local development and does not access real bank data. Common sandbox Link credentials are `user_good` / `pass_good`; if prompted for MFA, use `1234` or whatever is prompted.
- `production` uses real financial institutions and real user data. Production access requires the correct Plaid approvals and the production secret from the Dashboard. Switch both `PLAID_ENV=production` and `PLAID_SECRET=<your production secret>` together.
- This backend currently switches only on `sandbox` and `production`. As far as I can tell, `development` mode has been phased out.


### 4. Make the backend reachable for Plaid webhooks

Plaid needs a public HTTPS URL to send transaction webhooks back to your local backend. Expose the backend port, not the frontend port.

With the default config, the backend is on port `8080`.

#### Option A: ngrok

Install ngrok and connect your account:

- ngrok CLI quickstart: https://ngrok.com/docs/getting-started/

Then run:

```sh
ngrok http 8080
```

Copy the HTTPS forwarding URL that ngrok prints, for example:

```sh
WEBHOOK_URL=https://abc123.ngrok-free.app
```

If the app is already running, restart the backend after changing `.env` so the new webhook URL is used.

```sh
docker compose up -d --force-recreate backend
```

#### Option B: Cloudflare Quick Tunnel

Cloudflare Quick Tunnel can also expose a local port for development:

- Cloudflare Tunnel setup: https://developers.cloudflare.com/tunnel/setup/

Run:

```sh
cloudflared tunnel --url http://localhost:8080
```

Copy the generated `https://*.trycloudflare.com` URL into `WEBHOOK_URL`. If the app is already running, restart the backend.

```sh
docker compose up -d --force-recreate backend
```

### 5. Start the app

Once all the above has done to allow tunneling through ngrok or other, the plaid information is in your .env, you can build and start all services:

```sh
docker compose up --build
```

Or run in the background:

```sh
docker compose up --build -d
```

Open the app:

- Frontend: http://localhost:3000
- Backend API base: http://localhost:8080/api
- Postgres: `localhost:5432`

Useful commands:

```sh
docker compose ps
docker compose logs -f backend
docker compose logs -f frontend
docker compose down
```

`docker compose down` stops the containers but keeps the Postgres volume. To intentionally remove local database data too, run `docker compose down -v`.

## Local Architecture

Docker Compose starts three services:

- `db`: Postgres, using the `POSTGRES_*` and `DB_PORT` values.
- `backend`: Go API server, using the database, JWT, Plaid, CORS, and webhook env vars.
- `frontend`: Angular app served by nginx on `FRONTEND_PORT`.

The Plaid webhook endpoint is:

```text
POST /api/plaid/webhook
```

When `WEBHOOK_URL=https://abc123.ngrok-free.app`, the backend registers this webhook with Plaid as:

```text
https://abc123.ngrok-free.app/api/plaid/webhook
```
