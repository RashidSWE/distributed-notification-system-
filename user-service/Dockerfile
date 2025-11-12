# Base image
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files first (for layer caching)
COPY package*.json ./
RUN npm install

# Install dependencies (only production dependencies are often enough)
RUN npm ci --production 

# copy prisma folder
COPY prisma ./prisma

# Generate Prisma client
# This step is crucial and must happen *after* copying the files
RUN npx prisma generate

# Copy application source code
COPY . .

# Build the application for production (e.g., compile TypeScript to JavaScript)
RUN npm run build 

# Expose the Fastify port
EXPOSE 5000

# Start script:
# 1️⃣ Run Prisma migrations
# 2️⃣ Start the production server (e.g., node ./dist/index.js or npm start)
CMD sh -c "npx prisma migrate deploy && npm start"