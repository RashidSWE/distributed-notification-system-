import {create, login, getUserProfile, updateUser} from '../controllers/user.controller.js';
import {authenticate} from '../middlewares/auth.js';


export default async function userRoutes(app: any) {
  app.post('/api/users/register', create);
  app.post('/api/users/login', login);
  app.get('/api/users/profile', { preHandler: authenticate }, getUserProfile);
  app.put('/api/users/preferences', { preHandler: authenticate }, updateUser);
}