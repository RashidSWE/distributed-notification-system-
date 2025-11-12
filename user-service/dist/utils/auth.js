import bcrypt from "bcrypt";
import jwt from "jsonwebtoken";
export const hashPassword = async (password) => {
    return bcrypt.hash(password, 10);
};
export const comparePassword = async (password, hash) => {
    return bcrypt.compare(password, hash);
};
export const generateToken = (payload) => {
    return jwt.sign(payload, process.env.JWT_SECRET, { expiresIn: "1d" });
};
export const verifyToken = (token) => {
    return jwt.verify(token, process.env.JWT_SECRET);
};
//# sourceMappingURL=auth.js.map