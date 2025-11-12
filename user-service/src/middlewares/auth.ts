import { verifyToken } from '../utils/auth.js';
import type { FastifyRequest, FastifyReply } from 'fastify';

export const authenticate = async (req: FastifyRequest, reply: FastifyReply) => {
  try {
    const token = req.headers.authorization?.replace('Bearer ', '');
    console.log(token)
    if (!token) {
      return reply.code(401).send({ error: 'No token provided' });
    }
    
    const decoded = verifyToken(token);
    (req as any).user = decoded;
  } catch (err) {
    return reply.code(401).send({ error: 'Invalid token' });
  }
};