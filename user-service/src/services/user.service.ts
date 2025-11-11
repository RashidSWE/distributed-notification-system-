import { PrismaClient } from "@prisma/client";
import { hashPassword, comparePassword, generateToken } from "../utils/auth.js";

const prisma = new PrismaClient();

interface CreateUserInput {
  name: string;
  email: string;
  password: string;
  pushToken?: string;
  preferences?: { email: boolean; push: boolean };
}

export const createUser = async (data: CreateUserInput) => {
  const hashedPassword = await hashPassword(data.password);

  return prisma.user.create({
    data: {
      name: data.name,
      email: data.email,
      password: hashedPassword,
      ...(data.pushToken && { push_token: data.pushToken }),
      preference: {
        create: data.preferences || { email: true, push: true },
      },
    },
    include: { preference: true },
  });
};

export const loginUser = async (email: string, password: string) => {
  const user = await prisma.user.findUnique({ where: { email } });
  if (!user) throw new Error("Invalid credentials");

  const valid = await comparePassword(password, user.password);
  if (!valid) throw new Error("Invalid credentials");

  const token = generateToken({ id: user.id, email: user.email });
  return { user, token };
};

export const updatePreferences = async (userId: string, prefs: any) => {
  return prisma.preference.update({
    where: { user_id: userId },
    data: prefs,
  });
};

export const getUserById = async (userId: string) => {
  return prisma.user.findUnique({
    where: { id: userId },
    include: { preference: true },
  });
}


