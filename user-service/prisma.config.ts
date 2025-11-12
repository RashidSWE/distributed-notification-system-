import { defineConfig, env } from "prisma/config";
import "dotenv/config"; // Load environment variables from .env file

export default defineConfig({
  schema: "prisma/schema.prisma",
  migrations: {
    path: "prisma/migrations",
  },
  engine: "classic",
  datasource: {
    url: env("DATABASE_URL"),
  },
});
