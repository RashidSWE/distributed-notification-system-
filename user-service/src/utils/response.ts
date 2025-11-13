export interface PaginationMeta {
  total: number
  limit: number
  page: number
  total_pages: number
  has_next: boolean
  has_previous: boolean
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  message: string
  meta?: PaginationMeta
}

// Helper functions for consistent formatting
export const successResponse = <T>(
  data: T,
  message = "Success",
  meta?: PaginationMeta
): ApiResponse<T> => ({
  success: true,
  data,
  message,
  ...(meta ? { meta } : {}),
});

export const errorResponse = (
  message: string,
  error?: string
): ApiResponse<null> => ({
  success: false,
  message,
  ...(error ? { error } : {}),
});
