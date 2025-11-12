import { verifyToken } from '../utils/auth.js';
export const authenticate = async (req, reply) => {
    try {
        const token = req.headers.authorization?.replace('Bearer ', '');
        console.log(token);
        if (!token) {
            return reply.code(401).send({ error: 'No token provided' });
        }
        const decoded = verifyToken(token);
        req.user = decoded;
    }
    catch (err) {
        return reply.code(401).send({ error: 'Invalid token' });
    }
};
//# sourceMappingURL=auth.js.map