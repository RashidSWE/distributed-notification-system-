# Base image
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package.json and package-lock.json first for caching
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the app
COPY . .

# Expose app port
EXPOSE 5000

# Default command for local testing (uses tsx to run TypeScript directly)
CMD ["npm", "run", "dev"]
