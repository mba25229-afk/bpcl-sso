import { LayoutDashboard, TrendingUp, BarChart3, MessageSquare, HelpCircle, User, Menu, X } from 'lucide-react';
import logoImg from '../../imports/Official_BPCL_LOGO.jpg';

interface SidebarProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  isCollapsed: boolean;
  setIsCollapsed: (collapsed: boolean) => void;
}

const menuItems = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'analysis', label: 'Analysis', icon: BarChart3 },
  { id: 'trends', label: 'Trends', icon: TrendingUp },
  { id: 'feedback', label: 'Feedback', icon: MessageSquare },
  { id: 'support', label: 'Support', icon: HelpCircle },
  { id: 'profile', label: 'Profile', icon: User },
];

export function Sidebar({ activeTab, setActiveTab, isCollapsed, setIsCollapsed }: SidebarProps) {
  return (
    <div
      className={`h-full bg-[#007BC9] text-white transition-all duration-300 flex flex-col ${
        isCollapsed ? 'w-16' : 'w-64'
      }`}
    >
      <div className="p-4 flex items-center justify-between border-b border-white/10">
        {!isCollapsed && (
          <div className="flex items-center gap-3">
            <img src={logoImg} alt="BPCL Logo" className="w-14 h-14 object-contain bg-white rounded-lg p-1.5" />
            <div>
              <div className="font-bold text-base">BPCL</div>
              <div className="text-sm text-[#FFE000]">RetailIQ</div>
            </div>
          </div>
        )}
        {isCollapsed && (
          <img src={logoImg} alt="BPCL Logo" className="w-12 h-12 object-contain bg-white rounded p-1.5 mx-auto" />
        )}
        <button
          onClick={() => setIsCollapsed(!isCollapsed)}
          className="p-2 hover:bg-white/10 rounded-lg transition-colors"
        >
          {isCollapsed ? <Menu size={20} /> : <X size={20} />}
        </button>
      </div>

      <nav className="flex-1 p-2 space-y-1">
        {menuItems.map((item) => {
          const Icon = item.icon;
          const isActive = activeTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => setActiveTab(item.id)}
              className={`w-full flex items-center gap-3 px-3 py-3 rounded-lg transition-all ${
                isActive
                  ? 'bg-[#FFE000] text-[#007BC9]'
                  : 'text-white/80 hover:bg-white/10 hover:text-white'
              }`}
            >
              <Icon size={20} />
              {!isCollapsed && <span>{item.label}</span>}
            </button>
          );
        })}
      </nav>
    </div>
  );
}
