import { useState } from 'react';
import { Sidebar } from './components/Sidebar';
import { Dashboard } from './components/Dashboard';
import { Analysis } from './components/Analysis';
import { Trends } from './components/Trends';
import { Profile } from './components/Profile';
import { Feedback } from './components/Feedback';
import { Toaster } from 'sonner';

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [isCollapsed, setIsCollapsed] = useState(false);

  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard':
        return <Dashboard />;
      case 'analysis':
        return <Analysis />;
      case 'trends':
        return <Trends />;
      case 'feedback':
        return <Feedback />;
      case 'support':
        return (
          <div className="p-8">
            <h1 className="text-3xl font-bold text-gray-900">Support</h1>
            <p className="text-gray-600 mt-2">Support tickets and help resources</p>
          </div>
        );
      case 'profile':
        return <Profile />;
      default:
        return <Dashboard />;
    }
  };

  return (
    <>
      <Toaster position="top-right" richColors />
      <div className="size-full flex bg-gray-50">
        <Sidebar
          activeTab={activeTab}
          setActiveTab={setActiveTab}
          isCollapsed={isCollapsed}
          setIsCollapsed={setIsCollapsed}
        />
        <div className="flex-1 overflow-auto">
          {renderContent()}
        </div>
      </div>
    </>
  );
}