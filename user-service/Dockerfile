# Base image
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files first (for layer caching)
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy application source code
COPY . .

# Generate Prisma client
RUN npx prisma generate

# Expose the Fastify port
EXPOSE 5000

# Start script:
# 1️⃣ Wait for the database to be ready (optional, if you're using docker-compose depends_on it’s handled)
# 2️⃣ Run Prisma migrations
# 3️⃣ Start the dev server
CMD sh -c "npx prisma migrate deploy && npm run dev"
