import React from 'react';
import { NavLink } from 'react-router-dom';

const Sidebar = ({ navigationItems }) => {
  return (
    <div className="w-64 h-screen bg-gray-800 text-white">
      <div className="p-4 text-2xl font-bold">My App</div>
      <nav>
        <ul>
          {navigationItems.map((item) => (
            <li key={item.name}>
              <NavLink
                to={item.path}
                className={({ isActive }) =>
                  `block p-4 hover:bg-gray-700 ${isActive ? 'bg-gray-900' : ''}`
                }
              >
                {item.displayName}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
};

export default Sidebar; 