console.log("App starting...");

import Fastify from 'fastify';
import { PrismaClient } from '@prisma/client';
import userRoutes from './routes/user.routes.js';
import swaggerPlugin from './utils/swagger.js';

const app = Fastify({ logger: true });
const prisma = new PrismaClient();

// make prisma accessible via decorators
app.decorate('prisma', prisma);

// register plugins
await app.register(swaggerPlugin);

// register routes
await app.register(userRoutes);

// health check route
app.get('/api/health', async (req, reply) => {
  try {
    await prisma.user.count();
    return { status: 'ok', db: 'connected' };
  } catch (err) {
    if (err instanceof Error)
      return { status: 'error', db: 'disconnected', message: err.message };
  }
});

const start = async () => {
  try {
    await app.listen({ port: 5000, host: '0.0.0.0' });
    console.log('User Service running on port 5000');
    console.log('Swagger Docs available at http://localhost:5000/docs');
  } catch (err) {
    app.log.error(err);
    process.exit(1);
  }
};

start();
