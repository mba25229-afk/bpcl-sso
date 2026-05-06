import { User, Mail, Briefcase, MapPin, ChevronLeft } from 'lucide-react';

export function ManagerProfile({ onBack }: { onBack: () => void }) {
  const user = (() => {
    try {
      return JSON.parse(localStorage.getItem('bpcl_user') || 'null');
    } catch {
      return null;
    }
  })();

  const roleLabel: Record<string, string> = {
    admin: 'System Administrator',
    territory_manager: 'Territory Manager',
    ro_manager: 'RO Manager',
  };

  return (
    <div className="max-w-4xl mx-auto">
      <button
        onClick={onBack}
        className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-6 transition-colors"
      >
        <ChevronLeft size={20} />
        <span>Back to Dashboard</span>
      </button>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="h-32" style={{ backgroundColor: '#007BC9' }} />

        <div className="px-8 pb-8">
          <div className="flex items-end gap-6 -mt-16 mb-6">
            <div
              className="w-32 h-32 rounded-full border-4 border-white flex items-center justify-center"
              style={{ backgroundColor: '#FFE000' }}
            >
              <User size={64} className="text-gray-900" />
            </div>
            <div className="mb-4">
              <h2 className="text-2xl text-gray-900 mb-1">{user?.name ?? '—'}</h2>
              <p className="text-gray-600">{roleLabel[user?.role] ?? user?.role ?? '—'}</p>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-6 mb-6">
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center"
                  style={{ backgroundColor: '#007BC9' }}
                >
                  <Briefcase size={20} className="text-white" />
                </div>
                <div>
                  <p className="text-sm text-gray-600">Employee ID</p>
                  <p className="text-gray-900">{user?.employee_id ?? '—'}</p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center"
                  style={{ backgroundColor: '#007BC9' }}
                >
                  <Mail size={20} className="text-white" />
                </div>
                <div>
                  <p className="text-sm text-gray-600">Email</p>
                  <p className="text-gray-900">{user?.email ?? '—'}</p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center"
                  style={{ backgroundColor: '#007BC9' }}
                >
                  <MapPin size={20} className="text-white" />
                </div>
                <div>
                  <p className="text-sm text-gray-600">Territory</p>
                  <p className="text-gray-900">{user?.territory_code ?? 'All Territories'}</p>
                </div>
              </div>
            </div>

            <div className="space-y-4">
              <div className="bg-gray-50 rounded-lg p-4 space-y-3">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Role</span>
                  <span className="text-gray-900">{roleLabel[user?.role] ?? user?.role ?? '—'}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Account Status</span>
                  <span className="text-green-700 text-sm px-2 py-0.5 bg-green-50 rounded">
                    {user?.is_active ? 'Active' : 'Inactive'}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Last Login</span>
                  <span className="text-gray-900 text-sm">
                    {user?.last_login_at
                      ? new Date(user.last_login_at).toLocaleDateString('en-IN', {
                          day: '2-digit',
                          month: 'short',
                          year: 'numeric',
                        })
                      : '—'}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
