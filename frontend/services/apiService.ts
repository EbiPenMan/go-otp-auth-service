import type { User, PaginatedUsersResponse } from '../types';

// The base URL of your API, based on the swagger-doc.json file
const API_BASE_URL = 'http://localhost:8080';

/**
 * A helper function to handle fetch requests and errors.
 * @param url - The request URL
 * @param options - The fetch options
 * @returns - The JSON response
 */
const apiFetch = async (url: string, options: RequestInit = {}) => {
  const response = await fetch(url, options);

  if (!response.ok) {
    // Try to parse the error message from the response body
    const errorData = await response.json().catch(() => ({ error: 'An unknown error occurred' }));
    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
  }

  return response.json();
};

/**
 * Sends an OTP to the provided phone number.
 * @param phoneNumber - The phone number to send the OTP to.
 * @returns - A success message.
 */
export const sendOTP = (phoneNumber: string): Promise<{ message: string }> => {
  console.log(`[API REAL] Sending OTP to ${phoneNumber}.`);
  return apiFetch(`${API_BASE_URL}/otp/send`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ phone_number: phoneNumber }),
  });
};

/**
 * Verifies the OTP and returns a JWT token upon success.
 * @param phoneNumber - The user's phone number.
 * @param otp - The OTP code entered by the user.
 * @returns - An object containing the JWT token.
 */
export const verifyOTP = (phoneNumber: string, otp: string): Promise<{ token: string }> => {
  console.log(`[API REAL] Verifying OTP ${otp} for ${phoneNumber}.`);
  return apiFetch(`${API_BASE_URL}/otp/verify`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ phone_number: phoneNumber, otp: otp }),
  });
};

/**
 * Fetches a paginated and searchable list of users.
 * @param page - The page number.
 * @param limit - The number of items per page.
 * @param search - The search term for the phone number.
 * @param token - The JWT token for authentication.
 * @returns - The list of users and pagination information.
 */
export const getUsers = async (page: number, limit: number, search: string, token: string): Promise<PaginatedUsersResponse> => {
  console.log(`[API REAL] Fetching users. Page: ${page}, Limit: ${limit}, Search: "${search}"`);
  
  // Build the query string to send parameters
  const params = new URLSearchParams({
    page: String(page),
    limit: String(limit),
  });
  if (search) {
    params.append('search', search);
  }

  console.log('Token being attached to header:', token);
  console.log('Header value will be:', `Bearer ${token}`);

  // Your real API returns an object with a `data` key.
  // We'll map it to the `users` format your frontend expects.
  const response = await apiFetch(`${API_BASE_URL}/users?${params.toString()}`, {
    method: 'GET',
    headers: {
      // This part is correct, assuming the `token` variable is the raw string
      'Authorization': `Bearer ${token}`,
    },
  });

  // Adapt the API response to the structure expected by the frontend
  // API: { data: [user], total: number } -> Expected: { users: [user], total: number }
  const formattedUsers: User[] = response.data.map((user: any) => ({
    id: user.id,
    phoneNumber: user.phone_number,
    createdAt: user.created_at,
    // The `updatedAt` field is not in the API response definition, so it's omitted.
    // If your API adds it, you can map it here.
  }));

  return {
    users: formattedUsers,
    total: response.total,
    page,
    limit,
  };
};