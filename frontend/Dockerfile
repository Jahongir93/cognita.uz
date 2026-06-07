# ── Build stage ───────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM node:20-alpine

WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json /app/package-lock.json ./
RUN npm install --omit=dev

EXPOSE 3000
ENV PORT=3000
CMD ["node", "build"]
