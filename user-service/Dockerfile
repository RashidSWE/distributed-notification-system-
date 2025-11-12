# Base image
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files first (for layer caching)
COPY package*.json ./

# Install all dependencies (including dev dependencies for development)
RUN npm install

# Copy prisma schema
COPY prisma ./prisma

# Generate Prisma client
RUN npx prisma generate

# Copy application source code
COPY . .

# Expose the Fastify port
EXPOSE 5000

# Start script:
# 1️⃣ Run Prisma migrations
# 2️⃣ Start the development server
CMD sh -c "npx prisma migrate deploy && npm run dev"