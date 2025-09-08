import React, { useState, useEffect, useCallback } from 'react';
import type { User } from '../types';
import { getUsers } from '../services/apiService';
import Spinner from './Spinner';
import UserCard from './UserCard';
import { useAuth } from '../context/AuthContext'; // 1. Import useAuth

const UserDashboard: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalUsers, setTotalUsers] = useState(0);
  const [searchTerm, setSearchTerm] = useState('');
  
  const { token } = useAuth(); // 2. Get the token from the context

  const totalPages = Math.ceil(totalUsers / limit);

  const fetchUsers = useCallback(async () => {
    // 3. Add a guard clause: if there is no token, don't fetch.
    if (!token) {
      setError("Please log in to view the user list.");
      setIsLoading(false);
      setUsers([]); // Ensure user list is empty
      return;
    }

    setIsLoading(true);
    setError(null);
    try {
      // 4. Pass the token to the getUsers function
      const response = await getUsers(page, limit, searchTerm, token);
      setUsers(response.users);
      setTotalUsers(response.total);
    } catch (err) {
      // Handle expired token or other auth errors
      if (err instanceof Error && (err.message.includes('401') || err.message.toLowerCase().includes('unauthorized'))) {
         setError('Your session has expired. Please log out and log in again.');
      } else {
         setError(err instanceof Error ? err.message : 'Failed to fetch users.');
      }
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, searchTerm, token]); // 5. Add token to the dependency array

  useEffect(() => {
    const handler = setTimeout(() => {
        fetchUsers();
    }, 300); // Debounce search input

    return () => {
        clearTimeout(handler);
    };
  }, [fetchUsers]); // fetchUsers dependency is enough here

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
    setPage(1); // Reset to first page on new search
  };
  
  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages) {
      setPage(newPage);
    }
  };

  // The rest of your JSX remains the same
  return (
    <div className="space-y-6">
      <div className="p-6 bg-gray-800 rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-white">User Management</h1>
        <p className="mt-1 text-gray-400">Search, view, and manage registered users.</p>
        <div className="mt-4">
          <input
            type="text"
            placeholder="Search by phone number..."
            value={searchTerm}
            onChange={handleSearchChange}
            className="w-full max-w-sm px-4 py-2 text-white bg-gray-700 border border-gray-600 rounded-md placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center items-center h-64">
          <Spinner size="lg" />
        </div>
      ) : error ? (
        <div className="p-4 text-center text-red-300 bg-red-900 bg-opacity-50 rounded-lg">{error}</div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4">
            {users.map(user => <UserCard key={user.id} user={user} />)}
          </div>
          {users.length === 0 && !isLoading && (
            <div className="p-4 text-center text-gray-400 bg-gray-800 rounded-lg">No users found.</div>
          )}
          {totalPages > 1 && (
             <div className="flex items-center justify-between mt-6 px-2">
              <span className="text-sm text-gray-400">
                Page {page} of {totalPages}
              </span>
              <div className="flex items-center space-x-2">
                <button
                  onClick={() => handlePageChange(page - 1)}
                  disabled={page <= 1}
                  className="px-4 py-2 text-sm font-medium text-white bg-gray-700 rounded-md hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <button
                  onClick={() => handlePageChange(page + 1)}
                  disabled={page >= totalPages}
                  className="px-4 py-2 text-sm font-medium text-white bg-gray-700 rounded-md hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default UserDashboard;