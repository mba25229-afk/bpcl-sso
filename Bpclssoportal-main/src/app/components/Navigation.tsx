import { User, LogOut } from 'lucide-react';
import bpclLogo from '../../imports/Bharat_Petroleum-Logo.wine.png';

interface NavigationProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  onProfileClick: () => void;
  onLogout: () => void;
  showLeaderboard?: boolean;
  showTerritory?: boolean;
  showAdmin?: boolean;
}

export function Navigation({ activeTab, onTabChange, onProfileClick, onLogout, showLeaderboard, showTerritory, showAdmin }: NavigationProps) {
  const user = (() => {
    try {
      return JSON.parse(localStorage.getItem('bpcl_user') || 'null');
    } catch {
      return null;
    }
  })();

  const tabs = [
    { id: 'dashboard', label: 'Dashboard' },
    { id: 'analysis', label: 'Analysis' },
    ...(showTerritory ? [{ id: 'territory', label: 'Territory' }] : []),
    ...(showLeaderboard ? [{ id: 'leaderboard', label: 'Leaderboard' }] : []),
    ...(showAdmin ? [{ id: 'admin', label: 'Admin' }] : []),
  ];

  return (
    <nav className="bg-white border-b border-gray-200">
      <div className="max-w-[1440px] mx-auto px-8 py-4 flex items-center justify-between">
        <div className="flex items-center gap-12">
          <div className="flex items-center gap-3">
            <img src={bpclLogo} alt="BPCL Logo" className="h-12 w-auto" />
            <span className="font-semibold text-gray-900">BPCL Insight</span>
          </div>

          <div className="flex gap-8">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => onTabChange(tab.id)}
                className={`py-2 px-1 transition-colors ${
                  activeTab === tab.id
                    ? 'border-b-2 text-gray-900'
                    : 'text-gray-500 hover:text-gray-900'
                }`}
                style={activeTab === tab.id ? { borderColor: '#007BC9' } : {}}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        <div className="flex items-center gap-4">
          <button
            onClick={onProfileClick}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors cursor-pointer"
          >
            <User size={20} className="text-gray-600" />
            <span className="text-sm text-gray-700">{user?.name ?? 'Profile'}</span>
          </button>
          <button
            onClick={onLogout}
            className="p-2 rounded-lg hover:bg-gray-50 text-gray-600 hover:text-gray-900 transition-colors"
            title="Sign out"
          >
            <LogOut size={20} />
          </button>
        </div>
      </div>
    </nav>
  );
}
