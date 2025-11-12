import jwt from "jsonwebtoken";
export declare const hashPassword: (password: string) => Promise<string>;
export declare const comparePassword: (password: string, hash: string) => Promise<boolean>;
export declare const generateToken: (payload: Object) => string;
export declare const verifyToken: (token: string) => string | jwt.JwtPayload;
//# sourceMappingURL=auth.d.ts.map