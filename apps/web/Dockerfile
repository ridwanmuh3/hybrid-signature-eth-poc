# ── Build stage ──────────────────────────────────────────────────────
FROM node:20-slim AS builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

# Build args injected from docker-compose
ARG VITE_ORACLE_URL=http://localhost:9000
ARG VITE_CONTRACT_ADDRESS=
ARG VITE_ETHEREUM_RPC_URL=http://localhost:8540
ARG VITE_SEPOLIA_RPC_URL=

ENV VITE_ORACLE_URL=${VITE_ORACLE_URL}
ENV VITE_CONTRACT_ADDRESS=${VITE_CONTRACT_ADDRESS}
ENV VITE_ETHEREUM_RPC_URL=${VITE_ETHEREUM_RPC_URL}
ENV VITE_SEPOLIA_RPC_URL=${VITE_SEPOLIA_RPC_URL}

RUN pnpm build

# ── Runtime stage ────────────────────────────────────────────────────
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

# SPA fallback: serve index.html for all routes
RUN printf 'server {\n\
    listen 80;\n\
    root /usr/share/nginx/html;\n\
    index index.html;\n\
    location / {\n\
        try_files $uri $uri/ /index.html;\n\
    }\n\
}\n' > /etc/nginx/conf.d/default.conf

EXPOSE 80
