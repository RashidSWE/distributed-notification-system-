import { PrismaClient } from "@prisma/client";
import { hashPassword, comparePassword, generateToken } from "../utils/auth.js";
const prisma = new PrismaClient();
export const createUser = async (data) => {
    console.log("Creating user with data:", data);
    const hashedPassword = await hashPassword(data.password);
    return prisma.user.create({
        data: {
            name: data.name,
            email: data.email,
            password: hashedPassword,
            ...(data.push_token && { push_token: data.push_token }),
            preference: {
                create: data.preferences || { email: true, push: true },
            },
        },
        include: { preference: true },
    });
};
export const loginUser = async (email, password) => {
    const user = await prisma.user.findUnique({ where: { email } });
    if (!user)
        throw new Error("Invalid credentials");
    const valid = await comparePassword(password, user.password);
    if (!valid)
        throw new Error("Invalid credentials");
    const token = generateToken({ id: user.id, email: user.email });
    return { user, token };
};
export const updatePreferences = async (userId, prefs) => {
    prisma.preference.update({
        where: { user_id: userId },
        data: prefs,
    });
    return prisma.user.findUnique({ where: { id: userId }, include: { preference: true } });
};
export const updatePushToken = async (userId, push_token) => {
    return prisma.user.update({
        where: { id: userId },
        data: { push_token },
    });
};
export const getUserById = async (userId) => {
    return prisma.user.findUnique({
        where: { id: userId },
        include: { preference: true },
    });
};
//# sourceMappingURL=user.service.js.map