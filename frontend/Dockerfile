# --- Build stage ---
FROM node:26-alpine AS build
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
# VERSION is baked into the bundle via Vite (shown in the footer); defaults to
# "dev", the release workflow passes the git tag.
ARG VERSION=dev
ENV VITE_APP_VERSION=${VERSION}
RUN npm run build

# --- Serve stage ---
FROM nginx:alpine AS serve
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
