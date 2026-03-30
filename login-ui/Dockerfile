# ── Build stage ──
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
ENV PUBLIC_API_URL=http://localhost:5050
RUN npm run build
RUN npm prune --production

# ── Production stage ──
FROM node:22-alpine AS production

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=5053
ENV ORIGIN=http://localhost:5053

# Copy the built application
COPY --from=build /app/build ./build
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./

EXPOSE 5053

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:5053/login || exit 1

USER node

CMD ["node", "build/index.js"]
