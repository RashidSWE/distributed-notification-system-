import { loginUser, createUser, updatePreferences, getUserById, updatePushToken } from "../services/user.service.js";
import { successResponse, errorResponse } from "../utils/response.js";
import { PrismaClient } from "@prisma/client";
const prisma = new PrismaClient();
// app.post('/api/users/register', 
export const create = async (req, reply) => {
    try {
        const { name, email, password } = req.body;
        // Check if user exists
        const existingUser = await prisma.user.findUnique({ where: { email } });
        if (existingUser) {
            return reply.code(400).send(errorResponse('Registration failed', 'User already exists'));
        }
        const user = await createUser({ name, email, password });
        return reply.send(successResponse({ ...user }, "User registered successfully"));
    }
    catch (err) {
        return reply.code(500).send(errorResponse("Failed to register user", err instanceof Error ? err.message : "Unknown error"));
    }
};
// app.post('/api/users/login'
export const login = async (req, reply) => {
    try {
        const { email, password } = req.body;
        const { user, token } = await loginUser(email, password);
        return reply.send(successResponse({ ...user, token }, "Login successful"));
    }
    catch (err) {
        return reply.code(401).send(errorResponse("Login failed", err instanceof Error ? err.message : "Unknown error"));
    }
};
// app.get('/api/users/profile', { preHandler: authenticate }, 
export const getUserProfile = async (req, reply) => {
    try {
        const user = await getUserById(req.user.id);
        return reply.send(successResponse({ ...user }, "Profile retrieved successfully"));
    }
    catch (err) {
        return reply.code(500).send(errorResponse("Failed to retrieve profile", err instanceof Error ? err.message : "Unknown error"));
    }
};
// app.put('/api/users/preferences', { preHandler: authenticate }, 
export const updateUserPreferences = async (req, reply) => {
    try {
        const { email, push } = req.body;
        const user = await updatePreferences(req.user.id, { email, push });
        return reply.send(successResponse({ ...user }, "Preferences updated successfully"));
    }
    catch (err) {
        return reply.code(500).send(errorResponse("Failed to update preferences", err instanceof Error ? err.message : "Unknown error"));
    }
};
export const handlePushTokenUpdate = async (req, reply) => {
    try {
        const { push_token } = req.body;
        const user = await updatePushToken(req.user.id, push_token);
        return reply.send(successResponse({ ...user }, "Preferences updated successfully"));
    }
    catch (err) {
        return reply.code(500).send(errorResponse("Failed to update preferences", err instanceof Error ? err.message : "Unknown error"));
    }
};
//# sourceMappingURL=user.controller.js.map