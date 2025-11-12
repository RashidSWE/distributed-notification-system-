export interface PaginationMeta {
    total: number;
    limit: number;
    page: number;
    total_pages: number;
    has_next: boolean;
    has_previous: boolean;
}
export interface ApiResponse<T> {
    success: boolean;
    data?: T;
    error?: string;
    message: string;
    meta?: PaginationMeta;
}
export declare const successResponse: <T>(data: T, message?: string, meta?: PaginationMeta) => ApiResponse<T>;
export declare const errorResponse: (message: string, error?: string) => ApiResponse<null>;
//# sourceMappingURL=response.d.ts.map