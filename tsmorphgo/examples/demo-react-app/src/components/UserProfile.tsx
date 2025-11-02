import React from 'react';
import { User } from '../types';

interface UserProfileProps {
    user: User;
    onEdit: (user: User) => void;
    showAdvancedDetails?: boolean;
}

const UserProfile: React.FC<UserProfileProps> = ({ user, onEdit, showAdvancedDetails = false }) => {
    return (
        <div className="user-profile">
            <h2>{user.name}</h2>
            <p>{user.email}</p>
            {showAdvancedDetails && user.profile && (
                <div className="advanced-details">
                    <p>Bio: {user.profile.bio}</p>
                    <p>Website: {user.profile.website}</p>
                </div>
            )}
            <button onClick={() => onEdit(user)}>Edit</button>
        </div>
    );
};

export default UserProfile;
