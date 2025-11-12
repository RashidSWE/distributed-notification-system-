import type { FastifyInstance } from 'fastify';
import { create, login, getUserProfile, updateUserPreferences , handlePushTokenUpdate} from '../controllers/user.controller.js';
import { authenticate } from '../middlewares/auth.js';

export default async function userRoutes(app: FastifyInstance) {
  app.post('/api/users/register', create);
  app.post('/api/users/login', login);
  app.put('/api/users/me/push-token', { preHandler: authenticate }, handlePushTokenUpdate);
  app.get('/api/users/profile', { preHandler: authenticate }, getUserProfile);
  app.put('/api/users/profile', { preHandler: authenticate }, updateUserPreferences);
}