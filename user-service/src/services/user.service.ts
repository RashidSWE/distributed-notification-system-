import { PrismaClient } from "@prisma/client";
import { hashPassword, comparePassword, generateToken } from "../utils/auth.js";

const prisma = new PrismaClient();

interface CreateUserInput {
  name: string;
  email: string;
  password: string;
  push_token?: string;
  preferences?: { email: boolean; push: boolean };
}

type PreferenceUpdateData = {
  email?: boolean;
  push?: boolean;
};

export const createUser = async (data: CreateUserInput) => {
  console.log("Creating user with data:", data);
  const hashedPassword = await hashPassword(data.password);

  const user = await prisma.user.create({
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

  // Exclude password before returning
  const { password, ...userWithoutPassword } = user;
  return userWithoutPassword;
};

export const loginUser = async (email: string, password: string) => {
  const user = await prisma.user.findUnique({ where: { email } })
  if (!user) throw new Error("Invalid credentials");

  const valid = await comparePassword(password, user.password);
  if (!valid) throw new Error("Invalid credentials");

  const token = generateToken({ id: user.id, email: user.email, username: user.name });
  const { password: pwd, ...userWithoutPassword } = user;
  return { user: userWithoutPassword, token };
};

export const updatePreferences = async (userId: string, prefs: PreferenceUpdateData) => {
  await prisma.preference.update({
    where: { user_id: userId },
    data: prefs,
  });

  const user = await prisma.user.findUnique({ where: { id: userId }, include: { preference: true }});
  if (!user) throw new Error("User not found");

  const { password, ...userWithoutPassword } = user;
  return userWithoutPassword;
};

export const updatePushToken = async (userId: string, push_token: string) => {
  const user = await prisma.user.update({
    where: { id: userId },
    data: { push_token },
    include: { preference: true },
  });

  const { password, ...userWithoutPassword } = user;
  return userWithoutPassword;
};


export const getUserById = async (userId: string) => {
  return prisma.user.findUnique({
    where: { id: userId },
    select: {id: true, name: true, email: true, push_token: true, preference: true}
  });
}


