
import React from 'react';
import type { User } from '../types';

interface UserCardProps {
  user: User;
}

const UserCard: React.FC<UserCardProps> = ({ user }) => {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  return (
    <div className="bg-gray-800 p-4 rounded-lg shadow-md hover:bg-gray-700 transition-colors duration-200">
      <div className="flex items-center space-x-4">
        <div className="flex-shrink-0 h-12 w-12 rounded-full bg-indigo-500 flex items-center justify-center text-xl font-bold">
          {user.phoneNumber.slice(-2)}
        </div>
        <div>
          <p className="text-md font-semibold text-white">{user.phoneNumber}</p>
          <p className="text-sm text-gray-400">
            Registered on: {formatDate(user.createdAt)}
          </p>
        </div>
      </div>
    </div>
  );
};

export default UserCard;
