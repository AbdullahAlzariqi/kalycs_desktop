import React from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from '../components/Sidebar';

const Home = () => {
  const navigationItems = [
    { name: 'Dashboard', displayName: 'Dashboard', path: '/dashboard' },
    { name: 'Settings', displayName: 'Settings', path: '/settings' },
  ];

  return (
    <div className="flex">
      <Sidebar navigationItems={navigationItems} />
      <main className="flex-grow p-6">
        <Outlet />
      </main>
    </div>
  );
};

export default Home;
