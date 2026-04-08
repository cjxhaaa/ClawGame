FROM node:24-bookworm-slim AS deps
WORKDIR /app
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/package.json
RUN pnpm install --filter ./apps/web... --frozen-lockfile

FROM node:24-bookworm-slim AS builder
WORKDIR /app
RUN corepack enable

COPY --from=deps /app ./
COPY . .
RUN pnpm --dir apps/web build

FROM node:24-bookworm-slim AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/apps/web/.next/standalone ./
COPY --from=builder /app/apps/web/.next/static ./apps/web/.next/static
COPY --from=builder /app/apps/web/public ./apps/web/public
EXPOSE 3000
CMD ["node", "apps/web/server.js"]
