interface CreateUserInput {
    name: string;
    email: string;
    password: string;
    push_token?: string;
    preferences?: {
        email: boolean;
        push: boolean;
    };
}
type PreferenceUpdateData = {
    email?: boolean;
    push?: boolean;
};
export declare const createUser: (data: CreateUserInput) => Promise<{
    preference: {
        id: string;
        email: boolean;
        created_at: Date;
        updated_at: Date;
        push: boolean;
        user_id: string;
    } | null;
} & {
    id: string;
    name: string;
    email: string;
    password: string;
    push_token: string | null;
    created_at: Date;
    updated_at: Date;
}>;
export declare const loginUser: (email: string, password: string) => Promise<{
    user: {
        id: string;
        name: string;
        email: string;
        password: string;
        push_token: string | null;
        created_at: Date;
        updated_at: Date;
    };
    token: string;
}>;
export declare const updatePreferences: (userId: string, prefs: PreferenceUpdateData) => Promise<({
    preference: {
        id: string;
        email: boolean;
        created_at: Date;
        updated_at: Date;
        push: boolean;
        user_id: string;
    } | null;
} & {
    id: string;
    name: string;
    email: string;
    password: string;
    push_token: string | null;
    created_at: Date;
    updated_at: Date;
}) | null>;
export declare const updatePushToken: (userId: string, push_token: string) => Promise<{
    id: string;
    name: string;
    email: string;
    password: string;
    push_token: string | null;
    created_at: Date;
    updated_at: Date;
}>;
export declare const getUserById: (userId: string) => Promise<({
    preference: {
        id: string;
        email: boolean;
        created_at: Date;
        updated_at: Date;
        push: boolean;
        user_id: string;
    } | null;
} & {
    id: string;
    name: string;
    email: string;
    password: string;
    push_token: string | null;
    created_at: Date;
    updated_at: Date;
}) | null>;
export {};
//# sourceMappingURL=user.service.d.ts.map