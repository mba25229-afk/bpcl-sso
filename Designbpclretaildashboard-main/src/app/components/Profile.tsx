import { User, MapPin, Phone, Mail, Calendar, Award, TrendingUp, Building } from 'lucide-react';

export function Profile() {
  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Dealer Profile</h1>
        <p className="text-gray-600">Complete profile and business information</p>
      </div>

      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2 space-y-6">
          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Personal Information</h2>
            <div className="grid grid-cols-2 gap-6">
              <div>
                <label className="text-sm text-gray-600 mb-1 block">Full Name</label>
                <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <User size={18} className="text-[#007BC9]" />
                  <span className="text-gray-900">Mr. XYZ</span>
                </div>
              </div>
              <div>
                <label className="text-sm text-gray-600 mb-1 block">Dealer ID</label>
                <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <Award size={18} className="text-[#007BC9]" />
                  <span className="text-gray-900">BPCL-MH-2024-1542</span>
                </div>
              </div>
              <div>
                <label className="text-sm text-gray-600 mb-1 block">Email Address</label>
                <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <Mail size={18} className="text-[#007BC9]" />
                  <span className="text-gray-900">rajesh.sharma@bpcl.in</span>
                </div>
              </div>
              <div>
                <label className="text-sm text-gray-600 mb-1 block">Contact Number</label>
                <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <Phone size={18} className="text-[#007BC9]" />
                  <span className="text-gray-900">+91 98765 43210</span>
                </div>
              </div>
              <div>
                <label className="text-sm text-gray-600 mb-1 block">Date of Joining</label>
                <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <Calendar size={18} className="text-[#007BC9]" />
                  <span className="text-gray-900">March 15, 2020</span>
                </div>
              </div>
              <div>
                <label className="text-sm text-gray-600 mb-1 block">Experience</label>
                <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <TrendingUp size={18} className="text-[#007BC9]" />
                  <span className="text-gray-900">6 Years</span>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Retail Outlet Information</h2>
            <div className="space-y-4">
              <div>
                <label className="text-sm text-gray-600 mb-1 block">Outlet Name</label>
                <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <Building size={18} className="text-[#007BC9]" />
                  <span className="text-gray-900">BPCL Mumbai Central Retail Outlet</span>
                </div>
              </div>
              <div>
                <label className="text-sm text-gray-600 mb-1 block">Address</label>
                <div className="flex items-start gap-3 p-3 bg-gray-50 rounded-lg">
                  <MapPin size={18} className="text-[#007BC9] mt-1" />
                  <span className="text-gray-900">
                    Plot No. 42, Tardeo Road, Near Mumbai Central Station,<br />
                    Tardeo, Mumbai - 400034, Maharashtra, India
                  </span>
                </div>
              </div>
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="text-sm text-gray-600 mb-1 block">License Number</label>
                  <div className="p-3 bg-gray-50 rounded-lg">
                    <span className="text-gray-900">MH-PSL-2020-4578</span>
                  </div>
                </div>
                <div>
                  <label className="text-sm text-gray-600 mb-1 block">GST Number</label>
                  <div className="p-3 bg-gray-50 rounded-lg">
                    <span className="text-gray-900">27AABCU9603R1ZM</span>
                  </div>
                </div>
                <div>
                  <label className="text-sm text-gray-600 mb-1 block">PAN Number</label>
                  <div className="p-3 bg-gray-50 rounded-lg">
                    <span className="text-gray-900">AABCU9603R</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Outlet Facilities</h2>
            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  <span className="font-medium text-gray-900">Fuel Dispensers</span>
                </div>
                <p className="text-sm text-gray-600">8 Multi-Product Dispensers (MPD)</p>
              </div>
              <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  <span className="font-medium text-gray-900">BeCafe</span>
                </div>
                <p className="text-sm text-gray-600">Full-service café with seating</p>
              </div>
              <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  <span className="font-medium text-gray-900">In & Out Store</span>
                </div>
                <p className="text-sm text-gray-600">Convenience store operational</p>
              </div>
              <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  <span className="font-medium text-gray-900">UFill</span>
                </div>
                <p className="text-sm text-gray-600">Self-service fuel enabled</p>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-gradient-to-br from-[#007BC9] to-[#0056A3] rounded-xl p-6 text-white shadow-lg">
            <div className="text-center mb-4">
              <div className="w-24 h-24 bg-white rounded-full mx-auto mb-4 flex items-center justify-center">
                <User size={48} className="text-[#007BC9]" />
              </div>
              <h3 className="text-2xl font-bold">Mr.XYZ</h3>
              <p className="text-blue-100 mt-1">Retail Outlet Dealer</p>
            </div>
            <div className="mt-6 pt-6 border-t border-white/20">
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-blue-100">Current Rank</span>
                  <span className="font-bold text-[#FFE000]">#5</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-blue-100">Region</span>
                  <span className="font-bold">West Delhi</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-blue-100">Zone</span>
                  <span className="font-bold">Delhi NCR</span>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Performance Stats</h3>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm text-gray-600">Monthly Target</span>
                  <span className="font-bold text-[#007BC9]">91.8%</span>
                </div>
                <div className="w-full bg-gray-100 rounded-full h-2">
                  <div className="bg-[#007BC9] h-2 rounded-full" style={{ width: '91.8%' }}></div>
                </div>
              </div>
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm text-gray-600">Customer Satisfaction</span>
                  <span className="font-bold text-green-600">4.7/5.0</span>
                </div>
                <div className="w-full bg-gray-100 rounded-full h-2">
                  <div className="bg-green-600 h-2 rounded-full" style={{ width: '94%' }}></div>
                </div>
              </div>
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm text-gray-600">Non-Fuel Revenue</span>
                  <span className="font-bold text-[#FFE000]">20.0%</span>
                </div>
                <div className="w-full bg-gray-100 rounded-full h-2">
                  <div className="bg-[#FFE000] h-2 rounded-full" style={{ width: '20%' }}></div>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Certifications</h3>
            <div className="space-y-3">
              <div className="flex items-center gap-3 p-3 bg-green-50 rounded-lg">
                <Award size={18} className="text-green-600" />
                <div>
                  <p className="text-sm font-medium text-gray-900">HSE Certified</p>
                  <p className="text-xs text-gray-600">Valid till Dec 2026</p>
                </div>
              </div>
              <div className="flex items-center gap-3 p-3 bg-blue-50 rounded-lg">
                <Award size={18} className="text-blue-600" />
                <div>
                  <p className="text-sm font-medium text-gray-900">Quality Excellence</p>
                  <p className="text-xs text-gray-600">Awarded 2025</p>
                </div>
              </div>
              <div className="flex items-center gap-3 p-3 bg-yellow-50 rounded-lg">
                <Award size={18} className="text-yellow-600" />
                <div>
                  <p className="text-sm font-medium text-gray-900">ISO 9001:2015</p>
                  <p className="text-xs text-gray-600">Valid till Mar 2027</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
