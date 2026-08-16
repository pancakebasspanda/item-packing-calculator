FROM node:26-alpine AS builder

WORKDIR /app

# install dependencies first (caches this layer)
COPY package*.json ./
RUN npm ci

# Copy source code and build the Vite app
COPY . .
RUN npm run build


# serve with Nginx
FROM nginx:alpine

# copy the compiled Vite output (dist folder) to Nginx's web directory
COPY --from=builder /app/dist /usr/share/nginx/html

# expose port 80 inside the container
EXPOSE 80

# start Nginx
CMD ["nginx", "-g", "daemon off;"]