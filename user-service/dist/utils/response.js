// Helper functions for consistent formatting
export const successResponse = (data, message = "Success", meta) => ({
    success: true,
    data,
    message,
    ...(meta ? { meta } : {}),
});
export const errorResponse = (message, error) => ({
    success: false,
    message,
    ...(error ? { error } : {}),
});
//# sourceMappingURL=response.js.map