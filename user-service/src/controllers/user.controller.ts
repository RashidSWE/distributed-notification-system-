import type { FastifyReply } from "fastify";
import { loginUser, createUser, updatePreferences, getUserById} from "../services/user.service.js";
import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();


// app.post('/api/users/register', 
export const create = async (req: any, reply: FastifyReply) => {
  try {
    const { name, email, password } = req.body;
    
    // Check if user exists
    const existingUser = await prisma.user.findUnique({ where: { email } });
    if (existingUser) {
      return reply.code(400).send({
        success: false,
        error: 'User already exists',
        message: 'Registration failed'
      });
    }
    
    const user = await createUser({ name, email, password });
    
    return {
      success: true,
      data: { user: { ...user} },
      message: 'User registered successfully'
    };
  } catch (err) {
    return reply.code(500).send({
      success: false,
      error: err instanceof Error ? err.message : 'Unknown error',
      message: 'Registration failed'
    });
  }
};

// app.post('/api/users/login'
export const login = async (req: any, reply: FastifyReply) => {
  try {
    const { email, password } = req.body;
    
    const { user, token } = await loginUser(email, password);
    
    return {
      success: true,
      data: { user: { ...user, password: undefined }, token },
      message: 'Login successful'
    };
  } catch (err) {
    return reply.code(500).send({
      success: false,
      error: err instanceof Error ? err.message : 'Unknown error',
      message: 'Login failed'
    });
  }
};


// app.get('/api/users/profile', { preHandler: authenticate }, 
export const getUserProfile = async (req: any, reply: FastifyReply) => {
  try {
    const user = await getUserById(req.user.userId);
    return {success: true, data: { ...user }, message: 'Profile retrieved successfully'};
  } catch (err) {
    return reply.code(500).send({
      success: false,
      error: err instanceof Error ? err.message : 'Unknown error',
      message: 'Failed to retrieve profile'
    });
  }
}


// app.put('/api/users/preferences', { preHandler: authenticate }, 
export const updateUser = async (req: any, reply: FastifyReply) => {
  try {
    const { email, push, push_token } = req.body;
    
    const user = await updatePreferences(req.user.userId, { email, push, push_token });
    
    return {
      success: true,
      data: { ...user, password: undefined },
      message: 'Preferences updated successfully'
    };
  } catch (err) {
    return reply.code(500).send({
      success: false,
      error: err instanceof Error ? err.message : 'Unknown error',
      message: 'Failed to update preferences'
    });
  }
};

