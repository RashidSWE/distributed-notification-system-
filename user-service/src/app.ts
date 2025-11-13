console.log("🔥 App starting...");

import Fastify from 'fastify';
import { PrismaClient } from '@prisma/client';
import userRoutes from './routes/user.routes.js';


const app = Fastify({ logger: true });
const prisma = new PrismaClient();
const fastify = Fastify();

//
app.decorate('prisma', prisma);


await fastify.register(userRoutes);
await userRoutes(app);




// Health check
app.get('/api/health', async (req, reply) => {
  try {
    // Try a simple DB query
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
    console.log(`User Service running on port 5000`);
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
};

start();
