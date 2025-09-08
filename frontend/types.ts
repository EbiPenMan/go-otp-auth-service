
export interface User {
  id: string;
  phoneNumber: string;
  createdAt: string;
  updatedAt: string;
}

export interface PaginatedUsersResponse {
  users: User[];
  total: number;
  page: number;
  limit: number;
}
